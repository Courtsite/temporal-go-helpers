package channel

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type BasicReceiveWithTimeoutWorkflowResult struct {
	HasTimedOut bool
	Message     string
}

func BasicReceiveWithTimeoutWorkflow__WithPayload(ctx workflow.Context) (BasicReceiveWithTimeoutWorkflowResult, error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	childCtx, cancel := workflow.WithCancel(ctx)

	sigCh := workflow.GetSignalChannel(childCtx, "signal-receive-with-timeout")
	var message string
	hasTimedOut := ReceiveWithTimeout(ctx, sigCh, &message, time.Minute*30)

	if hasTimedOut {
		cancel()
	}

	var result BasicReceiveWithTimeoutWorkflowResult = BasicReceiveWithTimeoutWorkflowResult{
		HasTimedOut: hasTimedOut,
		Message:     message,
	}

	return result, nil
}

func BasicReceiveWithTimeoutWorkflow__NoPayload(ctx workflow.Context) (BasicReceiveWithTimeoutWorkflowResult, error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	childCtx, cancel := workflow.WithCancel(ctx)

	sigCh := workflow.GetSignalChannel(childCtx, "signal-receive-with-timeout")
	hasTimedOut := ReceiveWithTimeout(ctx, sigCh, nil, time.Minute*30)

	if hasTimedOut {
		cancel()
	}

	var result BasicReceiveWithTimeoutWorkflowResult = BasicReceiveWithTimeoutWorkflowResult{
		HasTimedOut: hasTimedOut,
		Message:     "",
	}

	return result, nil
}
