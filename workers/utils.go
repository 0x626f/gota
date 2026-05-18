package workers

import (
	"context"
	"fmt"
	"time"
)

// DoWithRetries calls task with attempt 0 once and retries up to retries times
// on error. Retry attempts are passed to task as 1..retries.
// Stops early if ctx is cancelled. Panics are caught and returned as errors.
func DoWithRetries(ctx context.Context, task ArgTask[int], retries int) (err error) {
	defer func() {
		if reason := recover(); reason != nil {
			err = fmt.Errorf("%v", reason)
		}
	}()

	err = task(0)
	for attempt := 1; err != nil && attempt <= retries; attempt++ {
		select {
		case <-ctx.Done():
			return err
		default:
			err = task(attempt)
		}
	}

	return err
}

// DoWithDelays calls task with delay 0 once and, on failure, retries after each
// provided delay in order. Each retry receives the delay that was waited before
// task execution. Stops early if ctx is cancelled or task succeeds. Panics are
// caught and returned as errors.
func DoWithDelays(ctx context.Context, task ArgTask[time.Duration], delays ...time.Duration) (err error) {
	defer func() {
		if reason := recover(); reason != nil {
			err = fmt.Errorf("%v", reason)
		}
	}()

	if err = task(0); err == nil {
		return err
	}

	timer := time.NewTimer(0)
	defer timer.Stop()

	for _, delay := range delays {
		timer.Reset(delay)
		select {
		case <-ctx.Done():
			return err
		case <-timer.C:
			if err = task(delay); err == nil {
				return err
			}
		}
	}

	return err
}

// Sleep blocks for delay, returning early if ctx is cancelled.
// Non-positive delays return immediately.
func Sleep(ctx context.Context, delay time.Duration) {
	if delay <= 0 {
		return
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return
	case <-ctx.Done():
		return
	}
}

// DoWithStopwatch executes task and returns how long it took alongside any
// error. Panics are caught and returned as errors.
func DoWithStopwatch(task Task) (d time.Duration, err error) {
	defer func() {
		if reason := recover(); reason != nil {
			err = fmt.Errorf("%v", reason)
		}
	}()

	start, err := time.Now(), task()
	d = time.Since(start)

	return
}
