package channel

import (
	"go.temporal.io/sdk/workflow"
)

func Drain(sigCh workflow.ReceiveChannel) int {
	total := 0
	for {
		hasData := sigCh.ReceiveAsync(nil)
		if !hasData {
			break
		}
		total += 1
	}
	return total
}
