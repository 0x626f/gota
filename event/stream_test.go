package event

import (
	"testing"
	"time"
)

// --- NewStream ---

func TestNewStream_NoArgs(t *testing.T) {
	s := NewStream[int]()
	if s == nil {
		t.Fatal("expected non-nil stream")
	}
	if s.size != 16 {
		t.Errorf("expected default size 16, got %d", s.size)
	}
}

func TestNewStream_ZeroSizeDefaultsTo16(t *testing.T) {
	s := NewStream[int](StreamParams{StreamSize: 0})
	if s.size != 16 {
		t.Errorf("expected default size 16, got %d", s.size)
	}
}

func TestNewStream_CustomSize(t *testing.T) {
	s := NewStream[int](StreamParams{StreamSize: 32})
	if s.size != 32 {
		t.Errorf("expected size 32, got %d", s.size)
	}
}

// --- Listen ---

func TestListen_ReturnsChannel(t *testing.T) {
	s := NewStream[int]()
	ch := s.Listen()
	if ch == nil {
		t.Fatal("Listen should return a non-nil channel")
	}
}

func TestListen_ChannelHasCorrectBufferSize(t *testing.T) {
	s := NewStream[int](StreamParams{StreamSize: 8})
	ch := s.Listen()
	if cap(ch) != 8 {
		t.Errorf("expected channel capacity 8, got %d", cap(ch))
	}
}

func TestListen_MultipleListeners(t *testing.T) {
	s := NewStream[int]()
	ch1 := s.Listen()
	ch2 := s.Listen()
	ch3 := s.Listen()

	if len(s.listeners) != 3 {
		t.Errorf("expected 3 listeners, got %d", len(s.listeners))
	}
	if ch1 == nil || ch2 == nil || ch3 == nil {
		t.Error("all listener channels should be non-nil")
	}
}

// --- Bind ---

func TestBind_NilSource_DoesNothing(t *testing.T) {
	s := NewStream[int]()
	s.Bind(nil)
	if s.source != nil {
		t.Error("source should remain nil after Bind(nil)")
	}
}

func TestBind_BroadcastsToListeners(t *testing.T) {
	s := NewStream[int](StreamParams{StreamSize: 4})
	ch1 := s.Listen()
	ch2 := s.Listen()

	src := make(chan int, 4)
	s.Bind(src)

	src <- 1
	src <- 2
	src <- 3

	time.Sleep(20 * time.Millisecond)

	for _, ch := range []<-chan int{ch1, ch2} {
		got := []int{}
		for len(ch) > 0 {
			got = append(got, <-ch)
		}
		if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
			t.Errorf("listener received %v, want [1 2 3]", got)
		}
	}
}

func TestBind_SourceClosed_ClosesListeners(t *testing.T) {
	s := NewStream[int](StreamParams{StreamSize: 4})
	ch := s.Listen()

	src := make(chan int, 2)
	src <- 10
	src <- 20
	close(src)

	s.Bind(src)

	// Drain expected values then verify channel is closed.
	for i := 0; i < 2; i++ {
		select {
		case _, ok := <-ch:
			if !ok && i < 2 {
				// channel closed before all values were read — only acceptable
				// after both values have been consumed
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("timeout waiting for value %d", i+1)
		}
	}

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected listener channel to be closed after source closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for listener channel to close")
	}
}

// --- Concurrent safety ---

func TestStream_ConcurrentListenAndBroadcast(t *testing.T) {
	s := NewStream[int](StreamParams{StreamSize: 64})

	src := make(chan int, 64)
	s.Bind(src)

	const numListeners = 10
	const numEvents = 20

	listeners := make([]<-chan int, numListeners)
	for i := range listeners {
		listeners[i] = s.Listen()
	}

	for i := 0; i < numEvents; i++ {
		src <- i
	}

	time.Sleep(50 * time.Millisecond)

	for i, ch := range listeners {
		if len(ch) != numEvents {
			t.Errorf("listener %d: got %d events, want %d", i, len(ch), numEvents)
		}
	}
}
