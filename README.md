# temporal-go-helpers

[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/courtsite/temporal-go-helpers)](https://pkg.go.dev/mod/github.com/courtsite/temporal-go-helpers)
[![CI](https://github.com/courtsite/temporal-go-helpers/workflows/CI/badge.svg)](https://github.com/Courtsite/temporal-go-helpers/actions?query=workflow%3ACI)


🔃 Common convenience methods, and developer ergonomics for [Temporal's](https://github.com/temporalio/temporal/) [Go SDK](https://github.com/temporalio/sdk-go).

_This is still under development. Use at your own risk._


## What's Inside?

- Saga
- Receive Signal with Timeout
- Drain Channel
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
    // `saga.Compensate` returns an error, you can check for it and/or log it.
    saga.Compensate(ctx)
    return err
}

// Both reservations succeeded.
```

It is worth noting that in simpler examples like above, we do not necessarily need or benefit from the Saga helper this module provides. We can simply call `workflow.ExecuteActivity(..)` within the `if err != nil` block. But, for more complex examples, it can become quite unmanageable, and in such cases, it is often easier to call `saga.Compensate(ctx)`. Your mileage may vary.

The compensation operations are executed in LIFO order.

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

s.AddFuture(timeout, func(f workflow.Future) {})
s.AddReceive(sigCh, func(c workflow.ReceiveChannel, more bool) {
    c.Receive(ctx, &signal)
})

s.Select(ctx)
```

But, with the helper:

```go
import "github.com/courtsite/temporal-go-helpers/channel"

sigCh := workflow.GetSignalChannel(ctx, "signal-with-timeout")
var signal SignalStruct
res := channel.ReceiveWithTimeout(ctx, sigCh, &signal, time.Minute * 30)

if res.IsCancelled {
    // Do something
}

if res.HasTimedOut {
    // Do something
}
```

It hides the need to set-up a timer, and selector, reducing cognitive overhead in your workflows.

It also handles context cancellation as well as potential race conditions with the timer future returning at the same time as the signal channel. In the latter scenario, the received signal will be favoured.


## Drain Channel

Based on https://community.temporal.io/t/continueasnew-signals/1008.

It is recommended that you drain your channel manually (i.e. do not use this library) if you must handle every signal.

This convenience method is simply for cases where you want to discard all signals. For example, you may have business logic that processes one signal at a time, but during the processing of that signal, you may not want to receive any other signals, perhaps to avoid some potential race conditions. If you do not have the ability to prevent signals from being sent during this time or if there is some likelihood for signals to be sent anyway, for safety, you can use `channel.Drain(ctx, sigCh)` to essentially "reset" the channel before waiting for newer signals to process.

If you are unsure about whether or not you should use this, then do not use it.

Example:

```go
import "github.com/courtsite/temporal-go-helpers/channel"

sigCh := workflow.GetSignalChannel(ctx, "signal")

for {
    // Reset, and ensure we only now accept new signals.
    // `n` will be the number of signals dropped.
    n := channel.Drain(ctx, sigCh)

    // Wait for, receive, and handle new signal from sigCh

    // Other business logic
}
```
