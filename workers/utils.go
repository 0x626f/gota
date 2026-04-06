package workers

import (
	"context"
	"fmt"
	"time"
)

// DoWithRetries calls task once and retries up to retries times on error.
// Stops early if ctx is cancelled. Panics are caught and returned as errors.
func DoWithRetries(ctx context.Context, task Task, retries int) (err error) {
	defer func() {
		if reason := recover(); reason != nil {
			err = fmt.Errorf("%v", reason)
		}
	}()

	err = task()
	for try := 0; err != nil && try < retries; try++ {
		select {
		case <-ctx.Done():
			return err
		default:
			err = task()
		}
	}

	return err
}

// DoWithDelays calls task once and, on failure, retries after each of the
// provided delays in order. Stops early if ctx is cancelled or task succeeds.
// Panics are caught and returned as errors.
func DoWithDelays(ctx context.Context, task Task, delays ...time.Duration) (err error) {
	defer func() {
		if reason := recover(); reason != nil {
			err = fmt.Errorf("%v", reason)
		}
	}()

	if err = task(); err == nil {
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
			if err = task(); err == nil {
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
