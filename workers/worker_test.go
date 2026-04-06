package workers

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"context"
)

// --- NewWorkerOnTicker ---

func TestNewWorkerOnTicker_ImmediateExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	executed := make(chan struct{}, 1)
	task := func() error {
		select {
		case executed <- struct{}{}:
		default:
		}
		return nil
	}

	w := NewWorkerOnTicker(ctx, task, 10*time.Second)
	w.Run()

	select {
	case <-executed:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("task was not executed immediately after Run()")
	}
}

func TestNewWorkerOnTicker_PeriodicExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count int32
	task := func() error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	w := NewWorkerOnTicker(ctx, task, 40*time.Millisecond)
	w.Run()
	time.Sleep(220 * time.Millisecond)

	// immediate + ~4 ticks in 220ms with 40ms interval
	if atomic.LoadInt32(&count) < 4 {
		t.Errorf("expected at least 4 executions, got %d", atomic.LoadInt32(&count))
	}
}

func TestNewWorkerOnTicker_StopsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var count int32
	task := func() error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	w := NewWorkerOnTicker(ctx, task, 20*time.Millisecond)
	w.Run()
	time.Sleep(80 * time.Millisecond)

	cancel()
	time.Sleep(30 * time.Millisecond)

	snapshot := atomic.LoadInt32(&count)
	time.Sleep(80 * time.Millisecond)

	if atomic.LoadInt32(&count) != snapshot {
		t.Errorf("task continued executing after context cancellation")
	}
}

func TestNewWorkerOnTicker_RunIsIdempotent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count int32
	task := func() error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	w := NewWorkerOnTicker(ctx, task, 10*time.Second)
	w.Run()
	w.Run()
	w.Run()

	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("expected exactly 1 execution from idempotent Run(), got %d", atomic.LoadInt32(&count))
	}
}

func TestNewWorkerOnTicker_OnErrorCalled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedErr := errors.New("task failed")
	errCh := make(chan error, 1)

	task := func() error { return expectedErr }

	w := NewWorkerOnTicker(ctx, task, 10*time.Second)
	w.OnError(func(err error) {
		select {
		case errCh <- err:
		default:
		}
	})
	w.Run()

	select {
	case err := <-errCh:
		if err != expectedErr {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("onError was not called after task returned error")
	}
}

func TestNewWorkerOnTicker_OnRecoveryCalled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recovered := make(chan any, 1)

	task := func() error {
		panic("unexpected failure")
	}

	w := NewWorkerOnTicker(ctx, task, 10*time.Second)
	w.OnRecovery(func(reason any) {
		select {
		case recovered <- reason:
		default:
		}
	})
	w.Run()

	select {
	case reason := <-recovered:
		if reason != "unexpected failure" {
			t.Errorf("unexpected recovery value: %v", reason)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("onRecovery was not called after task panicked")
	}
}

// --- NewWorkerOnSignal ---

func TestNewWorkerOnSignal_ExecutesOnSignal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	executed := make(chan struct{}, 1)
	task := func() error {
		select {
		case executed <- struct{}{}:
		default:
		}
		return nil
	}

	signal := make(chan struct{}, 1)
	w := NewWorkerOnSignal(ctx, task, signal)
	w.Run()

	signal <- struct{}{}

	select {
	case <-executed:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("task was not executed after signal")
	}
}

func TestNewWorkerOnSignal_MultipleSignals(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count int32
	task := func() error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	signal := make(chan struct{}, 5)
	w := NewWorkerOnSignal(ctx, task, signal)
	w.Run()

	for i := 0; i < 3; i++ {
		signal <- struct{}{}
	}
	time.Sleep(80 * time.Millisecond)

	if atomic.LoadInt32(&count) != 3 {
		t.Errorf("expected 3 executions, got %d", atomic.LoadInt32(&count))
	}
}

func TestNewWorkerOnSignal_StopsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var count int32
	task := func() error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	signal := make(chan struct{}, 10)
	w := NewWorkerOnSignal(ctx, task, signal)
	w.Run()

	signal <- struct{}{}
	time.Sleep(30 * time.Millisecond)

	cancel()
	time.Sleep(30 * time.Millisecond)

	snapshot := atomic.LoadInt32(&count)

	signal <- struct{}{}
	signal <- struct{}{}
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&count) != snapshot {
		t.Errorf("task executed after context cancellation")
	}
}

func TestNewWorkerOnSignal_ClosedChannelStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count int32
	task := func() error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	signal := make(chan struct{}, 2)
	w := NewWorkerOnSignal(ctx, task, signal)
	w.Run()

	signal <- struct{}{}
	time.Sleep(30 * time.Millisecond)

	close(signal)
	time.Sleep(30 * time.Millisecond)

	snapshot := atomic.LoadInt32(&count)
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&count) != snapshot {
		t.Errorf("task executed after signal channel was closed")
	}
}

func TestNewWorkerOnSignal_OnErrorCalled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedErr := errors.New("task failed")
	errCh := make(chan error, 1)

	task := func() error { return expectedErr }

	signal := make(chan struct{}, 1)
	w := NewWorkerOnSignal(ctx, task, signal)
	w.OnError(func(err error) {
		select {
		case errCh <- err:
		default:
		}
	})
	w.Run()

	signal <- struct{}{}

	select {
	case err := <-errCh:
		if err != expectedErr {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("onError was not called after task returned error")
	}
}

// --- NewWorkerOnEvent ---

func TestNewWorkerOnEvent_ExecutesWithPayload(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := make(chan string, 1)
	callback := func(val string) error {
		result <- val
		return nil
	}

	signal := make(chan string, 1)
	w := NewWorkerOnEvent(ctx, callback, signal)
	w.Run()

	signal <- "hello"

	select {
	case val := <-result:
		if val != "hello" {
			t.Errorf("expected %q, got %q", "hello", val)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("callback was not called after event")
	}
}

func TestNewWorkerOnEvent_MultipleEvents(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count int32
	callback := func(_ int) error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	signal := make(chan int, 10)
	w := NewWorkerOnEvent(ctx, callback, signal)
	w.Run()

	for i := 0; i < 5; i++ {
		signal <- i
	}
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&count) != 5 {
		t.Errorf("expected 5 executions, got %d", atomic.LoadInt32(&count))
	}
}

func TestNewWorkerOnEvent_StructPayload(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type Message struct {
		ID   int
		Text string
	}

	result := make(chan Message, 1)
	callback := func(msg Message) error {
		result <- msg
		return nil
	}

	signal := make(chan Message, 1)
	w := NewWorkerOnEvent(ctx, callback, signal)
	w.Run()

	expected := Message{ID: 42, Text: "hello"}
	signal <- expected

	select {
	case msg := <-result:
		if msg != expected {
			t.Errorf("expected %+v, got %+v", expected, msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("callback was not called after event")
	}
}

func TestNewWorkerOnEvent_StopsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var count int32
	callback := func(_ int) error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	signal := make(chan int, 10)
	w := NewWorkerOnEvent(ctx, callback, signal)
	w.Run()

	signal <- 1
	signal <- 2
	time.Sleep(50 * time.Millisecond)

	cancel()
	time.Sleep(30 * time.Millisecond)

	snapshot := atomic.LoadInt32(&count)

	signal <- 3
	signal <- 4
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&count) != snapshot {
		t.Errorf("callback executed after context cancellation")
	}
}

func TestNewWorkerOnEvent_ClosedChannelStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count int32
	callback := func(_ int) error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	signal := make(chan int, 2)
	w := NewWorkerOnEvent(ctx, callback, signal)
	w.Run()

	signal <- 1
	time.Sleep(30 * time.Millisecond)

	close(signal)
	time.Sleep(30 * time.Millisecond)

	snapshot := atomic.LoadInt32(&count)
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&count) != snapshot {
		t.Errorf("callback executed after event channel was closed")
	}
}

func TestNewWorkerOnEvent_OnErrorCalled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedErr := errors.New("callback failed")
	errCh := make(chan error, 1)

	callback := func(_ string) error { return expectedErr }

	signal := make(chan string, 1)
	w := NewWorkerOnEvent(ctx, callback, signal)
	w.OnError(func(err error) {
		select {
		case errCh <- err:
		default:
		}
	})
	w.Run()

	signal <- "trigger"

	select {
	case err := <-errCh:
		if err != expectedErr {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("onError was not called after callback returned error")
	}
}
