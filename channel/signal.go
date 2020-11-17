package channel

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func ReceiveWithTimeout(ctx workflow.Context, sigCh workflow.ReceiveChannel, valuePtr interface{}, timeout time.Duration) bool {
	selector := workflow.NewSelector(ctx)

	childCtx, cancel := workflow.WithCancel(ctx)
	defer cancel()

	selector.AddFuture(
		workflow.NewTimer(childCtx, timeout),
		func(f workflow.Future) {},
	)
	selector.AddReceive(
		sigCh,
		func(c workflow.ReceiveChannel, more bool) {},
	)

	selector.Select(ctx)

	var hasTimedOut bool
	if sigCh.ReceiveAsync(valuePtr) {
		hasTimedOut = false
	} else {
		hasTimedOut = true
	}

	return hasTimedOut
}
