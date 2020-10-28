package saga

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.temporal.io/sdk/workflow"
)

func BasicSagaWorkflow(ctx workflow.Context, initialAmount int) (int, error) {
	logger := workflow.GetLogger(ctx)

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	sagaCtx := New(ctx, SagaOptions{
		ParallelCompensation: false,
		ContinueWithError:    false,
	})

	ca := &CalculatorActivities{}

	currentAmount := initialAmount

	workflow.SetQueryHandler(ctx, "currentAmount", func(input []byte) (int, error) {
		return currentAmount, nil
	})

	err := workflow.ExecuteActivity(ctx, ca.Add, currentAmount, 10).Get(ctx, &currentAmount)
	if err != nil {
		logger.Error("activity failed", zap.Error(err))
		Compensate(sagaCtx)
		return 0, err
	}
	AddCompensation(sagaCtx, func(ctx workflow.Context) error {
		err := workflow.ExecuteActivity(ctx, ca.Minus, currentAmount, 5).Get(ctx, &currentAmount)
		if err != nil {
			logger.Error("compensation activity failed", zap.Error(err))
			return err
		}
		return nil
	})

	err = workflow.ExecuteActivity(ctx, ca.Add, currentAmount, 20).Get(ctx, &currentAmount)
	if err != nil {
		logger.Error("activity failed", zap.Error(err))
		Compensate(sagaCtx)
		return 0, err
	}
	AddCompensation(sagaCtx, func(ctx workflow.Context) error {
		err := workflow.ExecuteActivity(ctx, ca.Minus, currentAmount, 10).Get(ctx, &currentAmount)
		if err != nil {
			logger.Error("compensation activity failed", zap.Error(err))
			return err
		}
		return nil
	})

	err = workflow.ExecuteActivity(ctx, ca.Add, currentAmount, 30).Get(ctx, &currentAmount)
	if err != nil {
		logger.Error("activity failed", zap.Error(err))
		Compensate(sagaCtx)
		return 0, err
	}

	return currentAmount, nil
}

func MultipleCompensateSagaWorkflow(ctx workflow.Context, initialAmount int) (int, error) {
	logger := workflow.GetLogger(ctx)

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	sagaCtx := New(ctx, SagaOptions{
		ParallelCompensation: false,
		ContinueWithError:    false,
	})

	ca := &CalculatorActivities{}

	currentAmount := initialAmount

	workflow.SetQueryHandler(ctx, "currentAmount", func(input []byte) (int, error) {
		return currentAmount, nil
	})

	err := workflow.ExecuteActivity(ctx, ca.Add, currentAmount, 10).Get(ctx, &currentAmount)
	if err != nil {
		logger.Error("activity failed", zap.Error(err))
		Compensate(sagaCtx)
		return 0, err
	}
	AddCompensation(sagaCtx, func(ctx workflow.Context) error {
		err := workflow.ExecuteActivity(ctx, ca.Minus, currentAmount, 5).Get(ctx, &currentAmount)
		if err != nil {
			logger.Error("compensation activity failed", zap.Error(err))
			return err
		}
		return nil
	})

	Compensate(sagaCtx)
	Compensate(sagaCtx)

	return currentAmount, nil
}

type CalculatorActivities struct{}

func (ca *CalculatorActivities) Add(ctx context.Context, a int, b int) (int, error) {
	return a + b, nil
}

func (ca *CalculatorActivities) Minus(ctx context.Context, a int, b int) (int, error) {
	return a - b, nil
}
