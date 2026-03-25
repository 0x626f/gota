// Package workers provides utilities for running background tasks. It includes
// helpers to schedule work on a recurring interval, trigger work from a
// channel signal, and execute fallible operations with automatic retries or
// execution-time measurement.
package workers

import (
	"context"
	"time"
)

// RegisterWorkerOnDelay registers a task to run periodically at the specified delay interval.
// The task runs immediately once, then repeats at each tick of the delay duration.
// The goroutine will terminate when the context is cancelled.
//
// Parameters:
//   - ctx: Context for cancellation. When ctx.Done() is called, the task stops.
//   - callback: Function to execute on each tick. Should be idempotent.
//   - delay: Time duration between each execution after the initial call.
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	RegisterWorkerOnDelay(ctx, func() {
//	    fmt.Println("Task executed")
//	}, 5*time.Second)
//
// Note: This function starts a goroutine and returns immediately.
func RegisterWorkerOnDelay(ctx context.Context, callback func(), delay time.Duration) {
	go func() {
		ticker := time.NewTicker(delay)
		defer ticker.Stop()
		callback()
		for {
			select {
			case <-ticker.C:
				callback()
			case <-ctx.Done():
				return
			}
		}
	}()
}

// RegisterWorkerOnSignal registers a task to run whenever a signal is received on the provided channel.
// The task will execute each time a value is received from the signal channel.
// The goroutine will terminate when the context is cancelled.
//
// Parameters:
//   - ctx: Context for cancellation. When ctx.Done() is called, the task stops.
//   - callback: Function to execute when a signal is received.
//   - signal: Channel that triggers callback execution when values are received.
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	signal := make(chan any)
//	RegisterWorkerOnSignal(ctx, func() {
//	    fmt.Println("Signal received")
//	}, signal)
//	signal <- struct{}{} // Triggers the callback
//
// Note: This function starts a goroutine and returns immediately.
// If the signal channel is closed, the callback will not execute for that event.
func RegisterWorkerOnSignal(ctx context.Context, callback func(), signal <-chan any) {
	go func() {
		for {
			select {
			case _, ok := <-signal:
				if ok {
					callback()
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// RegisterWorkerOnEvent registers a task to run whenever a signal is received on the provided channel.
// Unlike RegisterWorkerOnSignal, this function passes the received value to the callback function.
// The goroutine will terminate when the context is cancelled.
//
// Type Parameters:
//   - T: The type of values sent through the signal channel and passed to the callback.
//
// Parameters:
//   - ctx: Context for cancellation. When ctx.Done() is called, the task stops.
//   - callback: Function to execute when a signal is received. Receives the signal value as an argument.
//   - signal: Channel that triggers callback execution when values are received.
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	signal := make(chan string)
//	RegisterWorkerOnEvent(ctx, func(msg string) {
//	    fmt.Printf("Received: %s\n", msg)
//	}, signal)
//	signal <- "Hello World" // Triggers the callback with "Hello World"
//
// Note: This function starts a goroutine and returns immediately.
// If the signal channel is closed, the callback will not execute for that event.
func RegisterWorkerOnEvent[T any](ctx context.Context, callback func(T), signal <-chan T) {
	go func() {
		for {
			select {
			case val, ok := <-signal:
				if ok {
					callback(val)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// DoWithRetries executes a function and retries it up to the specified number of times if it returns an error.
// The function is called immediately, and if it fails, it will be retried up to 'retries' times.
// The total number of attempts will be retries + 1 (initial attempt + retry attempts).
//
// Parameters:
//   - callback: Function to execute. Should return an error if it fails.
//   - retries: Maximum number of retry attempts after the initial call.
//
// Returns:
//   - error: The last error returned by callback, or nil if any attempt succeeded.
//
// Example:
//
//	err := DoWithRetries(func() error {
//	    return someOperation()
//	}, 3)
//	if err != nil {
//	    log.Printf("Operation failed after 4 attempts: %v", err)
//	}
//
// Note: This function blocks until either the callback succeeds or all retries are exhausted.
// There is no delay between retry attempts.
func DoWithRetries(callback func() error, retries int) error {
	err := callback()

	try := 1
	for err != nil && try <= retries {
		err = callback()
		try++
	}

	return err
}

// DoWithStopwatch executes a function and measures its execution time.
// It returns both the duration of execution and any error returned by the callback.
//
// Parameters:
//   - callback: Function to execute and measure. Can return an error.
//
// Returns:
//   - time.Duration: The time elapsed during callback execution.
//   - error: Any error returned by the callback, or nil if successful.
//
// Example:
//
//	duration, err := DoWithStopwatch(func() error {
//	    return performExpensiveOperation()
//	})
//	if err != nil {
//	    log.Printf("Operation failed after %v: %v", duration, err)
//	} else {
//	    log.Printf("Operation completed in %v", duration)
//	}
//
// Note: The timing starts immediately before callback execution and stops immediately after,
// regardless of whether the callback returns an error or succeeds.
func DoWithStopwatch(callback func() error) (time.Duration, error) {
	start, err := time.Now(), callback()
	duration := time.Since(start)
	return duration, err
}
