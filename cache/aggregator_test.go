package cache

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAggregator_CallReturnsCachedValue(t *testing.T) {
	store := NewPrimary[string, int]()
	aggregator := NewAggregator[int](store)

	var calls int32
	result, err := aggregator.Call("shape", func() (int, error) {
		atomic.AddInt32(&calls, 1)
		return 42, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 42 {
		t.Fatalf("expected 42, got %d", result)
	}

	result, err = aggregator.Call("shape", func() (int, error) {
		t.Fatal("action should not run on cache hit")
		return 0, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 42 {
		t.Fatalf("expected cached value 42, got %d", result)
	}
	if calls != 1 {
		t.Fatalf("expected action to run once, ran %d times", calls)
	}
}

func TestAggregator_CallDoesNotCacheErrors(t *testing.T) {
	store := NewPrimary[string, int]()
	aggregator := NewAggregator[int](store)
	expectedErr := errors.New("load failed")

	result, err := aggregator.Call("shape", func() (int, error) {
		return 10, expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected load error, got %v", err)
	}
	if result != 10 {
		t.Fatalf("expected failed result to be returned, got %d", result)
	}
	if cached, exists := store.Get("shape"); exists {
		t.Fatalf("failed result should not be cached, got %d", cached)
	}

	result, err = aggregator.Call("shape", func() (int, error) {
		return 20, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 20 {
		t.Fatalf("expected retry result 20, got %d", result)
	}
}

func TestAggregator_CallCachesSuccessfulResultWithTTL(t *testing.T) {
	store := NewPrimary[string, int]()
	aggregator := NewAggregator[int](store)

	result, err := aggregator.Call("shape", func() (int, error) {
		return 42, nil
	}, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 42 {
		t.Fatalf("expected 42, got %d", result)
	}

	if cached, exists := store.Get("shape"); !exists || cached != 42 {
		t.Fatalf("expected cached value 42 before TTL expiry, got %d, exists %t", cached, exists)
	}

	time.Sleep(20 * time.Millisecond)
	if cached, exists := store.Get("shape"); exists {
		t.Fatalf("expected cached value to expire, got %d", cached)
	}
}

func TestAggregator_CallDeduplicatesConcurrentSameKey(t *testing.T) {
	aggregator := NewAggregator[int](NewPrimary[string, int]())
	start := make(chan struct{})
	actionStarted := make(chan struct{})
	release := make(chan struct{})

	var calls int32
	const goroutines = 16
	results := make(chan int, goroutines)
	errs := make(chan error, goroutines)

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			result, err := aggregator.Call("shape", func() (int, error) {
				if atomic.AddInt32(&calls, 1) == 1 {
					close(actionStarted)
				}
				<-release
				return 42, nil
			})
			results <- result
			errs <- err
		}()
	}

	close(start)
	waitForChannel(t, actionStarted, "action to start")
	time.Sleep(10 * time.Millisecond)
	close(release)
	wg.Wait()
	close(results)
	close(errs)

	if calls != 1 {
		t.Fatalf("expected action to run once, ran %d times", calls)
	}
	for err := range errs {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	for result := range results {
		if result != 42 {
			t.Fatalf("expected shared result 42, got %d", result)
		}
	}
}

func TestAggregator_CallRunsDifferentKeysConcurrently(t *testing.T) {
	aggregator := NewAggregator[int]()
	startedA := make(chan struct{})
	startedB := make(chan struct{})
	release := make(chan struct{})
	errs := make(chan error, 2)

	go func() {
		result, err := aggregator.Call("a", func() (int, error) {
			close(startedA)
			<-release
			return 1, nil
		})
		if result != 1 {
			errs <- errors.New("unexpected result for key a")
			return
		}
		errs <- err
	}()

	go func() {
		result, err := aggregator.Call("b", func() (int, error) {
			close(startedB)
			<-release
			return 2, nil
		})
		if result != 2 {
			errs <- errors.New("unexpected result for key b")
			return
		}
		errs <- err
	}()

	waitForChannel(t, startedA, "action for key a to start")
	waitForChannel(t, startedB, "action for key b to start")
	close(release)

	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func waitForChannel(t *testing.T, ch <-chan struct{}, description string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", description)
	}
}
