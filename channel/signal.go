package channel

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type ReceiveWithTimeoutResult struct {
	HasTimedOut bool
	IsCancelled bool
}

func ReceiveWithTimeout(ctx workflow.Context, sigCh workflow.ReceiveChannel, valuePtr interface{}, timeout time.Duration) ReceiveWithTimeoutResult {
	selector := workflow.NewSelector(ctx)

	childCtx, cancel := workflow.WithCancel(ctx)
	defer cancel()

	timer := workflow.NewTimer(childCtx, timeout)

	selector.AddFuture(
		timer,
		func(f workflow.Future) {},
	)
	selector.AddReceive(
		sigCh,
		func(c workflow.ReceiveChannel, more bool) {},
	)

	selector.Select(ctx)

	var hasTimedOut bool
	var isCancelled bool

	// There can be 3 possible conditions:
	// 1. Signal received
	// 2. Timer fired because the context is "done"
	// 3. Timer fired

	if sigCh.ReceiveAsync(valuePtr) {
		hasTimedOut = false
		isCancelled = false
	} else if err := timer.Get(ctx, nil); err != nil && temporal.IsCanceledError(err) {
		hasTimedOut = false
		isCancelled = true
	} else {
		hasTimedOut = true
		isCancelled = false
	}

	return ReceiveWithTimeoutResult{
		HasTimedOut: hasTimedOut,
		IsCancelled: isCancelled,
	}
}
