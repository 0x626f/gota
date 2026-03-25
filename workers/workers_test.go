package workers

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestRegisterTaskWithDelay_ImmediateExecution verifies that the callback is executed immediately.
func TestRegisterTaskWithDelay_ImmediateExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	executed := make(chan struct{}, 1)
	callback := func() {
		select {
		case executed <- struct{}{}:
		default:
		}
	}

	RegisterWorkerOnDelay(ctx, callback, 100*time.Millisecond)

	select {
	case <-executed:
		// Success - immediate execution occurred
	case <-time.After(50 * time.Millisecond):
		t.Fatal("callback was not executed immediately")
	}
}

// TestRegisterTaskWithDelay_PeriodicExecution verifies that the callback is executed periodically.
func TestRegisterTaskWithDelay_PeriodicExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var counter int32
	callback := func() {
		atomic.AddInt32(&counter, 1)
	}

	delay := 50 * time.Millisecond
	RegisterWorkerOnDelay(ctx, callback, delay)

	// Wait for multiple executions
	time.Sleep(250 * time.Millisecond)

	count := atomic.LoadInt32(&counter)
	// Should have executed at least 4 times (immediate + 3-4 ticks in 250ms with 50ms delay)
	if count < 4 {
		t.Errorf("expected at least 4 executions, got %d", count)
	}
}

// TestRegisterTaskWithDelay_ContextCancellation verifies that the task stops when context is cancelled.
func TestRegisterTaskWithDelay_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var counter int32
	callback := func() {
		atomic.AddInt32(&counter, 1)
	}

	RegisterWorkerOnDelay(ctx, callback, 50*time.Millisecond)

	// Let it run for a bit
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Record count after cancellation
	countAfterCancel := atomic.LoadInt32(&counter)

	// Wait a bit more
	time.Sleep(150 * time.Millisecond)

	// Count should not have increased
	countAfterWait := atomic.LoadInt32(&counter)
	if countAfterWait > countAfterCancel {
		t.Errorf("task continued after context cancellation: before=%d, after=%d", countAfterCancel, countAfterWait)
	}
}

// TestRegisterTaskOnSignal_BasicExecution verifies that the callback executes on signal.
func TestRegisterTaskOnSignal_BasicExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	executed := make(chan struct{}, 1)
	signal := make(chan any, 1)

	callback := func() {
		executed <- struct{}{}
	}

	RegisterWorkerOnSignal(ctx, callback, signal)

	// Send a signal
	signal <- struct{}{}

	select {
	case <-executed:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("callback was not executed after signal")
	}
}

// TestRegisterTaskOnSignal_MultipleSignals verifies that the callback executes for each signal.
func TestRegisterTaskOnSignal_MultipleSignals(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var counter int32
	signal := make(chan any, 10)

	callback := func() {
		atomic.AddInt32(&counter, 1)
	}

	RegisterWorkerOnSignal(ctx, callback, signal)

	// Send multiple signals
	for i := 0; i < 5; i++ {
		signal <- struct{}{}
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	count := atomic.LoadInt32(&counter)
	if count != 5 {
		t.Errorf("expected 5 executions, got %d", count)
	}
}

// TestRegisterTaskOnSignal_ContextCancellation verifies that the task stops when context is cancelled.
func TestRegisterTaskOnSignal_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var counter int32
	signal := make(chan any, 10)

	callback := func() {
		atomic.AddInt32(&counter, 1)
	}

	RegisterWorkerOnSignal(ctx, callback, signal)

	// Send some signals
	signal <- struct{}{}
	signal <- struct{}{}
	time.Sleep(50 * time.Millisecond)

	cancel()

	// Try to send more signals after cancellation
	signal <- struct{}{}
	signal <- struct{}{}
	time.Sleep(50 * time.Millisecond)

	count := atomic.LoadInt32(&counter)
	if count != 2 {
		t.Errorf("expected exactly 2 executions before cancellation, got %d", count)
	}
}

// TestRegisterTaskOnSignal_ClosedChannel verifies that callback doesn't execute when channel is closed.
func TestRegisterTaskOnSignal_ClosedChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var counter int32
	signal := make(chan any, 1)

	callback := func() {
		atomic.AddInt32(&counter, 1)
	}

	RegisterWorkerOnSignal(ctx, callback, signal)

	// Send one signal
	signal <- struct{}{}
	time.Sleep(50 * time.Millisecond)

	// Close the channel
	close(signal)
	time.Sleep(50 * time.Millisecond)

	count := atomic.LoadInt32(&counter)
	if count != 1 {
		t.Errorf("expected exactly 1 execution, got %d", count)
	}
}

// TestRegisterTaskOnSignalWithValue_BasicExecution verifies that the callback receives the correct value.
func TestRegisterTaskOnSignalWithValue_BasicExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := make(chan string, 1)
	signal := make(chan string, 1)

	callback := func(val string) {
		result <- val
	}

	RegisterWorkerOnEvent(ctx, callback, signal)

	// Send a value
	signal <- "test-value"

	select {
	case val := <-result:
		if val != "test-value" {
			t.Errorf("expected 'test-value', got '%s'", val)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("callback was not executed after signal")
	}
}

// TestRegisterTaskOnSignalWithValue_MultipleValues verifies that the callback receives all values correctly.
func TestRegisterTaskOnSignalWithValue_MultipleValues(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	results := []int{}
	signal := make(chan int, 10)

	callback := func(val int) {
		mu.Lock()
		results = append(results, val)
		mu.Unlock()
	}

	RegisterWorkerOnEvent(ctx, callback, signal)

	// Send multiple values
	expected := []int{1, 2, 3, 4, 5}
	for _, v := range expected {
		signal <- v
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(results) != len(expected) {
		t.Fatalf("expected %d results, got %d", len(expected), len(results))
	}

	for i, exp := range expected {
		if results[i] != exp {
			t.Errorf("at index %d: expected %d, got %d", i, exp, results[i])
		}
	}
}

// TestRegisterTaskOnSignalWithValue_StructType verifies that the callback works with struct types.
func TestRegisterTaskOnSignalWithValue_StructType(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type Message struct {
		ID   int
		Text string
	}

	result := make(chan Message, 1)
	signal := make(chan Message, 1)

	callback := func(msg Message) {
		result <- msg
	}

	RegisterWorkerOnEvent(ctx, callback, signal)

	expected := Message{ID: 42, Text: "hello"}
	signal <- expected

	select {
	case msg := <-result:
		if msg.ID != expected.ID || msg.Text != expected.Text {
			t.Errorf("expected %+v, got %+v", expected, msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("callback was not executed after signal")
	}
}

// TestRegisterTaskOnSignalWithValue_ContextCancellation verifies that the task stops when context is cancelled.
func TestRegisterTaskOnSignalWithValue_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var counter int32
	signal := make(chan int, 10)

	callback := func(val int) {
		atomic.AddInt32(&counter, 1)
	}

	RegisterWorkerOnEvent(ctx, callback, signal)

	// Send some values
	signal <- 1
	signal <- 2
	time.Sleep(50 * time.Millisecond)

	cancel()

	// Try to send more values after cancellation
	signal <- 3
	signal <- 4
	time.Sleep(50 * time.Millisecond)

	count := atomic.LoadInt32(&counter)
	if count != 2 {
		t.Errorf("expected exactly 2 executions before cancellation, got %d", count)
	}
}

// TestCallWithRetries_SuccessFirstAttempt verifies that the function returns immediately on success.
func TestCallWithRetries_SuccessFirstAttempt(t *testing.T) {
	var attempts int32

	callback := func() error {
		atomic.AddInt32(&attempts, 1)
		return nil
	}

	err := DoWithRetries(callback, 3)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("expected 1 attempt, got %d", atomic.LoadInt32(&attempts))
	}
}

// TestCallWithRetries_SuccessAfterRetries verifies that the function succeeds after some failures.
func TestCallWithRetries_SuccessAfterRetries(t *testing.T) {
	var attempts int32

	callback := func() error {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	err := DoWithRetries(callback, 5)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", atomic.LoadInt32(&attempts))
	}
}

// TestCallWithRetries_ExhaustRetries verifies that the function returns error after exhausting retries.
func TestCallWithRetries_ExhaustRetries(t *testing.T) {
	var attempts int32
	expectedErr := errors.New("persistent error")

	callback := func() error {
		atomic.AddInt32(&attempts, 1)
		return expectedErr
	}

	err := DoWithRetries(callback, 3)

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	// Should be 1 initial + 3 retries = 4 total attempts
	if atomic.LoadInt32(&attempts) != 4 {
		t.Errorf("expected 4 attempts (1 initial + 3 retries), got %d", atomic.LoadInt32(&attempts))
	}
}

// TestCallWithRetries_ZeroRetries verifies that the function respects zero retries.
func TestCallWithRetries_ZeroRetries(t *testing.T) {
	var attempts int32
	expectedErr := errors.New("error")

	callback := func() error {
		atomic.AddInt32(&attempts, 1)
		return expectedErr
	}

	err := DoWithRetries(callback, 0)

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("expected 1 attempt, got %d", atomic.LoadInt32(&attempts))
	}
}

// TestCallWithRetries_NegativeRetries verifies that the function handles negative retries.
func TestCallWithRetries_NegativeRetries(t *testing.T) {
	var attempts int32
	expectedErr := errors.New("error")

	callback := func() error {
		atomic.AddInt32(&attempts, 1)
		return expectedErr
	}

	err := DoWithRetries(callback, -1)

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("expected 1 attempt, got %d", atomic.LoadInt32(&attempts))
	}
}

// TestDoWithStopwatch_SuccessfulExecution verifies that duration is measured correctly for successful execution.
func TestDoWithStopwatch_SuccessfulExecution(t *testing.T) {
	expectedDelay := 100 * time.Millisecond

	callback := func() error {
		time.Sleep(expectedDelay)
		return nil
	}

	duration, err := DoWithStopwatch(callback)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Allow 20ms tolerance for timing variations
	tolerance := 20 * time.Millisecond
	if duration < expectedDelay || duration > expectedDelay+tolerance {
		t.Errorf("expected duration around %v, got %v", expectedDelay, duration)
	}
}

// TestDoWithStopwatch_ErrorExecution verifies that duration is measured even when callback returns an error.
func TestDoWithStopwatch_ErrorExecution(t *testing.T) {
	expectedDelay := 50 * time.Millisecond
	expectedErr := errors.New("test error")

	callback := func() error {
		time.Sleep(expectedDelay)
		return expectedErr
	}

	duration, err := DoWithStopwatch(callback)

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	// Verify duration was still measured despite the error
	tolerance := 20 * time.Millisecond
	if duration < expectedDelay || duration > expectedDelay+tolerance {
		t.Errorf("expected duration around %v, got %v", expectedDelay, duration)
	}
}

// TestDoWithStopwatch_FastExecution verifies that very fast operations are measured correctly.
func TestDoWithStopwatch_FastExecution(t *testing.T) {
	callback := func() error {
		return nil
	}

	duration, err := DoWithStopwatch(callback)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Fast operations should complete in under 10ms
	if duration > 10*time.Millisecond {
		t.Errorf("expected fast execution under 10ms, got %v", duration)
	}

	if duration < 0 {
		t.Errorf("duration should not be negative, got %v", duration)
	}
}

// TestDoWithStopwatch_MediumExecution verifies timing for medium-duration operations.
func TestDoWithStopwatch_MediumExecution(t *testing.T) {
	expectedDelay := 200 * time.Millisecond

	callback := func() error {
		time.Sleep(expectedDelay)
		return nil
	}

	duration, err := DoWithStopwatch(callback)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	tolerance := 30 * time.Millisecond
	if duration < expectedDelay || duration > expectedDelay+tolerance {
		t.Errorf("expected duration around %v, got %v", expectedDelay, duration)
	}
}

// TestDoWithStopwatch_NilError verifies that nil error is returned correctly.
func TestDoWithStopwatch_NilError(t *testing.T) {
	callback := func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	_, err := DoWithStopwatch(callback)

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// TestDoWithStopwatch_MultipleExecutions verifies consistent behavior across multiple calls.
func TestDoWithStopwatch_MultipleExecutions(t *testing.T) {
	expectedDelay := 50 * time.Millisecond

	callback := func() error {
		time.Sleep(expectedDelay)
		return nil
	}

	tolerance := 20 * time.Millisecond

	for i := 0; i < 3; i++ {
		duration, err := DoWithStopwatch(callback)

		if err != nil {
			t.Errorf("iteration %d: expected no error, got %v", i, err)
		}

		if duration < expectedDelay || duration > expectedDelay+tolerance {
			t.Errorf("iteration %d: expected duration around %v, got %v", i, expectedDelay, duration)
		}
	}
}
