package channel

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type BasicReceiveWithTimeoutWorkflowResult struct {
	HasTimedOut bool
	Message     string
}

func BasicReceiveWithTimeoutWorkflow(ctx workflow.Context) (BasicReceiveWithTimeoutWorkflowResult, error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	sigCh := workflow.GetSignalChannel(ctx, "signal-receive-with-timeout")
	var message string
	hasTimedOut := ReceiveWithTimeout(ctx, sigCh, &message, time.Minute*30)

	var result BasicReceiveWithTimeoutWorkflowResult = BasicReceiveWithTimeoutWorkflowResult{
		HasTimedOut: hasTimedOut,
		Message:     message,
	}

	return result, nil
}
