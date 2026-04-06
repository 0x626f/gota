package workers

import (
	"context"
	"errors"
	"testing"
	"time"
)

// --- DoWithRetries ---

func TestDoWithRetries_SuccessOnFirstAttempt(t *testing.T) {
	ctx := context.Background()
	var calls int

	err := DoWithRetries(ctx, func() error {
		calls++
		return nil
	}, 3)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected task called once, got %d", calls)
	}
}

func TestDoWithRetries_SuccessAfterRetry(t *testing.T) {
	ctx := context.Background()
	var calls int

	err := DoWithRetries(ctx, func() error {
		calls++
		if calls < 3 {
			return errors.New("not yet")
		}
		return nil
	}, 5)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestDoWithRetries_ExhaustRetries(t *testing.T) {
	ctx := context.Background()
	var calls int
	expectedErr := errors.New("persistent error")

	err := DoWithRetries(ctx, func() error {
		calls++
		return expectedErr
	}, 3)

	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
	// 1 initial call + 3 retries
	if calls != 4 {
		t.Errorf("expected 4 calls (1 initial + 3 retries), got %d", calls)
	}
}

func TestDoWithRetries_ZeroRetries(t *testing.T) {
	ctx := context.Background()
	var calls int
	expectedErr := errors.New("error")

	err := DoWithRetries(ctx, func() error {
		calls++
		return expectedErr
	}, 0)

	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
	if calls != 1 {
		t.Errorf("expected task called exactly once with zero retries, got %d", calls)
	}
}

func TestDoWithRetries_StopsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var calls int
	expectedErr := errors.New("fail")

	err := DoWithRetries(ctx, func() error {
		calls++
		if calls == 2 {
			cancel()
		}
		return expectedErr
	}, 10)

	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
	if calls >= 10 {
		t.Errorf("expected retries to stop on context cancellation, but ran all 10 retries")
	}
}

func TestDoWithRetries_PanicRecovery(t *testing.T) {
	ctx := context.Background()

	err := DoWithRetries(ctx, func() error {
		panic("unexpected panic")
	}, 0)

	if err == nil {
		t.Fatal("expected error from panic recovery, got nil")
	}
}

// --- DoWithDelays ---

func TestDoWithDelays_SuccessOnFirstCall(t *testing.T) {
	ctx := context.Background()
	var calls int

	err := DoWithDelays(ctx, func() error {
		calls++
		return nil
	}, 100*time.Millisecond, 200*time.Millisecond)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected task called once on immediate success, got %d", calls)
	}
}

func TestDoWithDelays_RetriesAfterDelay(t *testing.T) {
	ctx := context.Background()
	var calls int

	err := DoWithDelays(ctx, func() error {
		calls++
		if calls < 3 {
			return errors.New("not yet")
		}
		return nil
	}, 20*time.Millisecond, 20*time.Millisecond, 20*time.Millisecond)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestDoWithDelays_ExhaustDelays(t *testing.T) {
	ctx := context.Background()
	var calls int
	expectedErr := errors.New("persistent error")

	err := DoWithDelays(ctx, func() error {
		calls++
		return expectedErr
	}, 10*time.Millisecond, 10*time.Millisecond)

	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
	// 1 initial + 1 per delay
	if calls != 3 {
		t.Errorf("expected 3 calls (1 initial + 2 retries), got %d", calls)
	}
}

func TestDoWithDelays_NoDelays(t *testing.T) {
	ctx := context.Background()
	var calls int
	expectedErr := errors.New("error")

	err := DoWithDelays(ctx, func() error {
		calls++
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
	if calls != 1 {
		t.Errorf("expected task called once when no delays provided, got %d", calls)
	}
}

func TestDoWithDelays_StopsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls int

	err := DoWithDelays(ctx, func() error {
		calls++
		cancel() // cancel during first call so the delay wait is skipped
		return errors.New("fail")
	}, 10*time.Second)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 1 {
		t.Errorf("expected 1 call before context cancellation interrupted delay, got %d", calls)
	}
}

func TestDoWithDelays_PanicRecovery(t *testing.T) {
	ctx := context.Background()

	err := DoWithDelays(ctx, func() error {
		panic("boom")
	})

	if err == nil {
		t.Fatal("expected error from panic recovery, got nil")
	}
}

// --- DoWithStopwatch ---

func TestDoWithStopwatch_MeasuresDuration(t *testing.T) {
	delay := 80 * time.Millisecond
	tolerance := 30 * time.Millisecond

	duration, err := DoWithStopwatch(func() error {
		time.Sleep(delay)
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if duration < delay || duration > delay+tolerance {
		t.Errorf("expected duration ~%v, got %v", delay, duration)
	}
}

func TestDoWithStopwatch_ReturnsError(t *testing.T) {
	expectedErr := errors.New("task failed")

	_, err := DoWithStopwatch(func() error {
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestDoWithStopwatch_MeasuresDurationOnError(t *testing.T) {
	delay := 50 * time.Millisecond
	tolerance := 30 * time.Millisecond

	duration, err := DoWithStopwatch(func() error {
		time.Sleep(delay)
		return errors.New("fail")
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if duration < delay || duration > delay+tolerance {
		t.Errorf("expected duration ~%v even on error, got %v", delay, duration)
	}
}

func TestDoWithStopwatch_PanicRecovery(t *testing.T) {
	_, err := DoWithStopwatch(func() error {
		panic("stopwatch panic")
	})

	if err == nil {
		t.Fatal("expected error from panic recovery, got nil")
	}
}

// --- Sleep ---

func TestSleep_SleepsForDuration(t *testing.T) {
	ctx := context.Background()
	delay := 80 * time.Millisecond
	tolerance := 30 * time.Millisecond

	start := time.Now()
	Sleep(ctx, delay)
	elapsed := time.Since(start)

	if elapsed < delay || elapsed > delay+tolerance {
		t.Errorf("expected sleep of ~%v, got %v", delay, elapsed)
	}
}

func TestSleep_InterruptedByContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	Sleep(ctx, 10*time.Second)
	elapsed := time.Since(start)

	if elapsed > 100*time.Millisecond {
		t.Errorf("expected sleep interrupted by cancellation within ~30ms, took %v", elapsed)
	}
}

func TestSleep_ZeroDelayReturnsImmediately(t *testing.T) {
	ctx := context.Background()

	start := time.Now()
	Sleep(ctx, 0)

	if time.Since(start) > 10*time.Millisecond {
		t.Errorf("expected immediate return for zero delay")
	}
}

func TestSleep_NegativeDelayReturnsImmediately(t *testing.T) {
	ctx := context.Background()

	start := time.Now()
	Sleep(ctx, -time.Second)

	if time.Since(start) > 10*time.Millisecond {
		t.Errorf("expected immediate return for negative delay")
	}
}
