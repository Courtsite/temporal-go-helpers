package channel

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func ReceiveWithTimeout(ctx workflow.Context, sigCh workflow.ReceiveChannel, valuePtr interface{}, timeout time.Duration) bool {
	selector := workflow.NewSelector(ctx)

	var hasTimedOut bool

	selector.AddFuture(
		workflow.NewTimer(ctx, timeout),
		func(f workflow.Future) {
			hasTimedOut = true
		},
	)
	selector.AddReceive(
		sigCh,
		func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, valuePtr)
		},
	)

	selector.Select(ctx)

	return hasTimedOut
}
