package workers

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultPoolSize(t *testing.T) {
	expected := runtime.NumCPU() - 1
	if expected < 1 {
		expected = 1
	}

	if got := DefaultPoolSize(); got != expected {
		t.Errorf("expected %d workers, got %d", expected, got)
	}
}

func TestNewPool_UsesDefaultWorkerCount(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := NewPool(ctx, PoolParams[int]{
		Callback: func(_ int) error { return nil },
	})

	if got := pool.WorkerCount(); got != DefaultPoolSize() {
		t.Errorf("expected %d workers, got %d", DefaultPoolSize(), got)
	}
	if got := pool.QueueSize(); got != pool.WorkerCount() {
		t.Errorf("expected queue size %d, got %d", pool.WorkerCount(), got)
	}
}

func TestNewPool_UsesConfiguredQueueSize(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := NewPool(ctx, PoolParams[int]{
		Callback:  func(_ int) error { return nil },
		Workers:   2,
		QueueSize: 8,
	})

	if got := pool.QueueSize(); got != 8 {
		t.Errorf("expected queue size 8, got %d", got)
	}
	if got := cap(pool.Queue()); got != 8 {
		t.Errorf("expected queue capacity 8, got %d", got)
	}
}

func TestNewPool_ProcessesQueueFromChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := make(chan int, 3)
	pool := NewPool(ctx, PoolParams[int]{
		Callback: func(value int) error {
			result <- value
			return nil
		},
		Workers: 2,
	})
	pool.Run()

	queue := pool.Queue()
	queue <- 1
	queue <- 2
	queue <- 3

	seen := map[int]bool{}
	for len(seen) < 3 {
		select {
		case value := <-result:
			seen[value] = true
		case <-time.After(100 * time.Millisecond):
			t.Fatal("pool did not process submitted queue items")
		}
	}
}

func TestNewPool_ProcessesQueueConcurrently(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	started := make(chan struct{}, 2)
	release := make(chan struct{})
	pool := NewPool(ctx, PoolParams[int]{
		Callback: func(_ int) error {
			started <- struct{}{}
			<-release
			return nil
		},
		Workers: 2,
	})
	pool.Run()

	queue := pool.Queue()
	go func() { queue <- 1 }()
	go func() { queue <- 2 }()

	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(100 * time.Millisecond):
			close(release)
			t.Fatal("expected both workers to process queue items concurrently")
		}
	}

	close(release)
}

func TestPool_CloseAndWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count int32
	pool := NewPool(ctx, PoolParams[int]{
		Callback: func(_ int) error {
			atomic.AddInt32(&count, 1)
			return nil
		},
		Workers: 2,
	})
	pool.Run()

	queue := pool.Queue()
	queue <- 1
	queue <- 2
	pool.Close()

	done := make(chan struct{})
	go func() {
		pool.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("pool did not stop after queue channel was closed")
	}

	if got := atomic.LoadInt32(&count); got != 2 {
		t.Errorf("expected 2 processed queue items, got %d", got)
	}
}

func TestPool_StopsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	pool := NewPool(ctx, PoolParams[int]{
		Callback: func(_ int) error { return nil },
		Workers:  2,
	})
	pool.Run()
	cancel()

	done := make(chan struct{})
	go func() {
		pool.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("pool did not stop after context cancellation")
	}
}

func TestPool_OnErrorCalled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedErr := errors.New("job failed")
	errCh := make(chan error, 1)
	pool := NewPool(ctx, PoolParams[int]{
		Callback: func(_ int) error {
			return expectedErr
		},
		Workers: 1,
	})
	returned := pool.OnError(func(err error) {
		errCh <- err
	})
	if returned != pool {
		t.Fatal("OnError did not return the pool instance")
	}
	pool.Run()

	pool.Queue() <- 1

	select {
	case err := <-errCh:
		if err != expectedErr {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("onError was not called after callback returned error")
	}
}

func TestPool_OnRecoveryCalled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recovered := make(chan any, 1)
	pool := NewPool(ctx, PoolParams[int]{
		Callback: func(_ int) error {
			panic("unexpected failure")
		},
		Workers: 1,
	})
	returned := pool.OnRecovery(func(reason any) {
		recovered <- reason
	})
	if returned != pool {
		t.Fatal("OnRecovery did not return the pool instance")
	}
	pool.Run()

	pool.Queue() <- 1

	select {
	case reason := <-recovered:
		if reason != "unexpected failure" {
			t.Errorf("unexpected recovery value: %v", reason)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("onRecovery was not called after callback panicked")
	}
}
