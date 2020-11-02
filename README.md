# temporal-go-helpers

[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/courtsite/temporal-go-helpers)](https://pkg.go.dev/mod/github.com/courtsite/temporal-go-helpers)


🔃 Common convenience methods, and developer ergonomics for [Temporal's](https://github.com/temporalio/temporal/) [Go SDK](https://github.com/temporalio/sdk-go).

_This is still under development. Use at your own risk._


## What's Inside?

- Saga
- Receive Signal with Timeout
- Any suggestions? Ping us!


## Saga

Based on https://github.com/uber-go/cadence-client/issues/797.

The [Saga Pattern](https://microservices.io/patterns/data/saga.html) is used for managing data consistency across microservices in distributed transaction scenarios.

This module provides APIs for easily executing compensation rollback logic, even after the parent workflow has been cancelled.

Example:

```go
// Modified from https://docs.temporal.io/docs/use-cases-distributed-transactions/

import "github.com/courtsite/temporal-go-helpers/saga"

ctx = saga.New(ctx, saga.SagaOptions{
    ParallelCompensation: false,
    ContinueWithError:    false,
})

// Step 1: Book a hotel.
var hotelResID string
err := workflow.ExecuteActivity(ctx, activities.BookHotel, name).Get(ctx, &hotelResID)
if err != nil {
    // If the hotel reservation failed, fail the Workflow.
    return err
}
saga.AddCompensation(ctx, func(ctx workflow.Context) error {
    cancelErr := workflow.ExecuteActivity(ctx, activities.CancelHotel, hotelResID)
    if cancelErr != nil {
        // The hotel cancellation failed... probably, some manual action is needed.
        return cancelErr
    }
    return nil
}

// Step 2: Book a flight.
var flightResID string
err := workflow.ExecuteActivity(ctx, activities.BookFlight, name).Get(ctx, &flightResID)
if err != nil {
    // If the flight reservation failed, cancel the hotel.
    saga.Compensate(ctx)
    return err
}

// Both reservations succeeded.
```

It is worth noting that in simpler examples like above, we do not necessarily need or benefit from the Saga helper this module provides. We can simply call `workflow.ExecuteActivity(..)` within the `if err != nil` block. But, for more complex examples, it can become quite unmanageable, and in such cases, it is often easier to call `saga.Compensate(ctx)`. Your mileage may vary.

The various compensation rollback logic can be executed in parallel by setting `ParallelCompensation` to `true`.


## Receive Signal with Timeout

Based on https://github.com/uber-go/cadence-client/issues/789.

It is fairly common to want to wait on a signal with a timeout. For example, we may want to wait for a user to click on a verification token link or wait for an OTP to be entered, but timeout, and continue if it takes too long.

With the vanilla Go SDK, this pattern will often look like this:

```go
sigCh := workflow.GetSignalChannel(ctx, "signal-with-timeout")
timeout := workflow.NewTimer(ctx, time.Minute * 30)

var signal SignalStruct

s := workflow.NewSelector(ctx)

s.AddFuture(timeout, func(f Future) {})
s.AddReceive(sigCh, func(c Channel, more bool) {
    c.Receive(ctx, &signal)
})

s.Select(ctx)
```

But, with the helper:

```go
import "github.com/courtsite/temporal-go-helpers/channel"

sigCh := workflow.GetSignalChannel(ctx, "signal-with-timeout")
var signal SignalStruct
hasTimedOut := channel.ReceiveWithTimeout(ctx, sigCh, &signal, time.Minute * 30)

if hasTimedOut {
    // Do something
}
```

It hides the need to set-up a timer, and selector, reducing cognitive overhead in your workflows.
