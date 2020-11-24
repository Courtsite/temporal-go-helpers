package saga

import (
	"fmt"
	"strings"
	"sync"

	"go.temporal.io/sdk/workflow"
)

// A unique key for identifying the saga state in context.
// This is to avoid collisions with keys defined in other packages.
// See https://golang.org/pkg/context/#Context
type key int

var sagaStateContextKey key

type Context workflow.Context

type CompensationError struct {
	Errors []error
}

func (e *CompensationError) AddError(err error) {
	e.Errors = append(e.Errors, err)
}

func (e *CompensationError) HasErrors() bool {
	return len(e.Errors) > 0
}

func (e *CompensationError) Error() string {
	if !e.HasErrors() {
		return "no errors in saga compensation"
	}

	var errors []string
	for _, err := range e.Errors {
		errors = append(errors, err.Error())
	}
	return fmt.Sprintf("error(s) in saga compensation: %s", strings.Join(errors, "; "))
}

type SagaOptions struct {
	ParallelCompensation bool
	ContinueWithError    bool
}

type compensationOp func(ctx workflow.Context) error

type sagaState struct {
	ctx     Context
	options SagaOptions

	compensationOps []compensationOp
	cancelFunc      workflow.CancelFunc

	once sync.Once
}

func New(ctx workflow.Context, options SagaOptions) Context {
	sagaCtx, cancelFunc := workflow.NewDisconnectedContext(ctx)
	sagaState := &sagaState{
		ctx:        sagaCtx,
		options:    options,
		cancelFunc: cancelFunc,
	}
	ctxWithSaga := workflow.WithValue(ctx, sagaStateContextKey, sagaState)
	return ctxWithSaga
}

func AddCompensation(ctx Context, op compensationOp) {
	sagaState := mustGetSagaState(ctx)
	sagaState.compensationOps = append(sagaState.compensationOps, op)
}

func Compensate(ctx Context) error {
	var err error
	var isCalled bool

	sagaState := mustGetSagaState(ctx)

	sagaState.once.Do(func() {
		err = compensate(sagaState.ctx, sagaState)
		isCalled = true
	})

	if err != nil {
		return err
	}

	if !isCalled {
		panic("Compensate: Compensate should only be called once")
	}

	return nil
}

func compensate(ctx Context, sagaState *sagaState) error {
	if sagaState.options.ParallelCompensation {
		var futures []workflow.Future

		for _, op := range sagaState.compensationOps {
			future, settable := workflow.NewFuture(ctx)
			workflow.Go(ctx, func(wfctx workflow.Context) {
				err := op(wfctx)
				if err != nil {
					settable.SetError(err)
					return
				}
			})
			futures = append(futures, future)
		}

		compensationError := &CompensationError{}

		for _, future := range futures {
			err := future.Get(ctx, nil)
			if err != nil {
				compensationError.AddError(err)
			}
		}

		if compensationError.HasErrors() {
			return compensationError
		}
	} else {
		for i := len(sagaState.compensationOps) - 1; i >= 0; i-- {
			op := sagaState.compensationOps[i]
			err := op(ctx)
			if err != nil && !sagaState.options.ContinueWithError {
				return err
			}
		}
	}

	return nil
}

func Cancel(ctx Context) {
	sagaState := mustGetSagaState(ctx)
	sagaState.cancelFunc()
}

func mustGetSagaState(ctx Context) *sagaState {
	sagaState, ok := ctx.Value(sagaStateContextKey).(*sagaState)
	if !ok || sagaState == nil {
		panic("mustGetSagaState: Not a saga context")
	}

	return sagaState
}
