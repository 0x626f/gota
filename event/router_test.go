package event

import (
	"errors"
	"sync"
	"testing"
)

// testEvent is a minimal Event implementation for tests.
type testEvent struct {
	id   ID
	data any
}

func (e testEvent) Id() ID    { return e.id }
func (e testEvent) Data() any { return e.data }

// --- NewRouter ---

func TestNewRouter_NoArgs(t *testing.T) {
	r := NewRouter()
	if r == nil {
		t.Fatal("expected non-nil router")
	}
	if r.routes == nil {
		t.Error("routes map should be initialized")
	}
	if r.async {
		t.Error("async should default to false")
	}
	if r.onError != nil {
		t.Error("onError should default to nil")
	}
}

func TestNewRouter_WithParams(t *testing.T) {
	called := false
	errHandler := func(Event, error) { called = true }

	r := NewRouter(RouterParams{Async: true, OnErr: errHandler})
	if !r.async {
		t.Error("async should be true")
	}
	if r.onError == nil {
		t.Fatal("onError should be set")
	}
	r.onError(nil, nil)
	if !called {
		t.Error("onError was not set correctly")
	}
}

// --- OnEvent ---

func TestOnEvent_RegistersHandler(t *testing.T) {
	r := NewRouter()
	r.OnEvent("foo", func(any) error { return nil })
	if _, ok := r.routes["foo"]; !ok {
		t.Error("handler not registered for 'foo'")
	}
}

func TestOnEvent_OverwritesHandler(t *testing.T) {
	r := NewRouter()
	r.OnEvent("foo", func(any) error { return errors.New("first") })
	r.OnEvent("foo", func(any) error { return errors.New("second") })

	var got error
	r.onError = func(_ Event, err error) { got = err }
	r.Route(testEvent{id: "foo"})

	if got == nil || got.Error() != "second" {
		t.Errorf("expected 'second' error, got %v", got)
	}
}

// --- Route (sync) ---

func TestRoute_CallsRegisteredHandler(t *testing.T) {
	r := NewRouter()

	var received any
	r.OnEvent("ping", func(v any) error {
		received = v
		return nil
	})

	ev := testEvent{id: "ping", data: "payload"}
	r.Route(ev)

	if received != ev.data {
		t.Errorf("handler received %v, want %v", received, ev)
	}
}

func TestRoute_NoHandler_CallsOnError(t *testing.T) {
	var gotEvent Event
	var gotErr error

	r := NewRouter(RouterParams{
		OnErr: func(e Event, err error) {
			gotEvent = e
			gotErr = err
		},
	})

	ev := testEvent{id: "unregistered"}
	r.Route(ev)

	if gotEvent != ev {
		t.Errorf("onError received wrong event: %v", gotEvent)
	}
	if gotErr == nil {
		t.Error("expected an error from onError, got nil")
	}
}

func TestRoute_NoHandler_NoOnError_NoPanic(t *testing.T) {
	r := NewRouter()
	r.Route(testEvent{id: "ghost"}) // must not panic
}

func TestRoute_HandlerError_CallsOnError(t *testing.T) {
	handlerErr := errors.New("handler failed")

	var gotErr error
	r := NewRouter(RouterParams{
		OnErr: func(_ Event, err error) { gotErr = err },
	})
	r.OnEvent("boom", func(any) error { return handlerErr })

	r.Route(testEvent{id: "boom"})

	if gotErr != handlerErr {
		t.Errorf("expected handler error, got %v", gotErr)
	}
}

func TestRoute_HandlerError_NoOnError_NoPanic(t *testing.T) {
	r := NewRouter()
	r.OnEvent("boom", func(any) error { return errors.New("oops") })
	r.Route(testEvent{id: "boom"}) // must not panic
}

func TestRoute_HandlerSuccess_OnErrorNotCalled(t *testing.T) {
	called := false
	r := NewRouter(RouterParams{
		OnErr: func(Event, error) { called = true },
	})
	r.OnEvent("ok", func(any) error { return nil })
	r.Route(testEvent{id: "ok"})

	if called {
		t.Error("onError should not be called on handler success")
	}
}

// --- Route (async) ---

func TestRoute_Async_CallsHandler(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	r := NewRouter(RouterParams{Async: true})
	r.OnEvent("a", func(any) error {
		defer wg.Done()
		return nil
	})

	r.Route(testEvent{id: "a"})
	wg.Wait()
}

func TestRoute_Async_HandlerError_CallsOnError(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	var gotErr error
	r := NewRouter(RouterParams{
		Async: true,
		OnErr: func(_ Event, err error) {
			gotErr = err
			wg.Done()
		},
	})
	r.OnEvent("bad", func(any) error { return errors.New("async err") })

	r.Route(testEvent{id: "bad"})
	wg.Wait()

	if gotErr == nil || gotErr.Error() != "async err" {
		t.Errorf("unexpected error: %v", gotErr)
	}
}

// --- Multiple events ---

func TestRoute_MultipleEvents(t *testing.T) {
	r := NewRouter()

	calls := map[string]int{}
	for _, id := range []string{"a", "b", "c"} {
		id := id
		r.OnEvent(id, func(any) error {
			calls[id]++
			return nil
		})
	}

	r.Route(testEvent{id: "a"})
	r.Route(testEvent{id: "b"})
	r.Route(testEvent{id: "b"})
	r.Route(testEvent{id: "c"})

	if calls["a"] != 1 || calls["b"] != 2 || calls["c"] != 1 {
		t.Errorf("unexpected call counts: %v", calls)
	}
}
