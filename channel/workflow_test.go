package channel

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type BasicReceiveWithTimeoutWorkflowResult struct {
	HasTimedOut bool
	IsCancelled bool
	Message     string
}

func BasicReceiveWithTimeoutWorkflow__WithPayload(ctx workflow.Context) (BasicReceiveWithTimeoutWorkflowResult, error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// This is to ensure the signal channel gets cleaned up after timing out.
	childCtx, cancel := workflow.WithCancel(ctx)

	sigCh := workflow.GetSignalChannel(childCtx, "signal-receive-with-timeout")
	var message string
	res := ReceiveWithTimeout(ctx, sigCh, &message, time.Minute*30)

	if res.HasTimedOut {
		cancel()
	}

	var result BasicReceiveWithTimeoutWorkflowResult = BasicReceiveWithTimeoutWorkflowResult{
		HasTimedOut: res.HasTimedOut,
		IsCancelled: res.IsCancelled,
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

	// This is to ensure the signal channel gets cleaned up after timing out.
	childCtx, cancel := workflow.WithCancel(ctx)

	sigCh := workflow.GetSignalChannel(childCtx, "signal-receive-with-timeout")
	res := ReceiveWithTimeout(ctx, sigCh, nil, time.Minute*30)

	if res.HasTimedOut {
		cancel()
	}

	var result BasicReceiveWithTimeoutWorkflowResult = BasicReceiveWithTimeoutWorkflowResult{
		HasTimedOut: res.HasTimedOut,
		IsCancelled: res.IsCancelled,
		Message:     "",
	}

	return result, nil
}

func BasicDrainWorkflow(ctx workflow.Context) (int, error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	sigCh := workflow.GetSignalChannel(ctx, "signal")

	totalDrained := 0

	for {
		totalDrained += Drain(sigCh)

		var message string
		sigCh.Receive(ctx, &message)

		if message == "OK" {
			// Do some processing
			workflow.Sleep(ctx, time.Second*5)
		} else if message == "STOP" {
			break
		}
	}

	return totalDrained, nil
}
