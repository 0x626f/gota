package event

import (
	"sync"
)

// Stream broadcasts values from a single source channel to multiple listeners.
// It is safe for concurrent use.
type Stream[T any] struct {
	source    <-chan T
	listeners []chan<- T
	size      int

	mu sync.RWMutex
}

// StreamParams configures a Stream.
type StreamParams struct {
	// StreamSize sets the buffer capacity of each listener channel.
	// Defaults to 16 when zero.
	StreamSize int
}

// NewStream creates a new Stream. An optional StreamParams may be provided to
// configure listener buffer size; the default buffer size is 16.
func NewStream[T any](params ...StreamParams) *Stream[T] {
	var p StreamParams
	if len(params) > 0 {
		p = params[0]
	}

	if p.StreamSize == 0 {
		p.StreamSize = 16
	}

	return &Stream[T]{size: p.StreamSize}
}

// Listen registers a new listener and returns its receive channel. The channel
// is buffered according to StreamParams.StreamSize. It is safe to call Listen
// concurrently with Bind.
func (stream *Stream[T]) Listen() <-chan T {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	listener := make(chan T, stream.size)
	stream.listeners = append(stream.listeners, listener)

	return listener
}

// Bind attaches source as the stream's input and starts broadcasting its values
// to all registered listeners. If source is nil, Bind is a no-op. When source
// is closed all listener channels are closed and the internal goroutine exits.
// Bind should be called at most once per Stream.
func (stream *Stream[T]) Bind(source <-chan T) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	if source == nil {
		return
	}

	stream.source = source

	go func() {
		for {
			select {
			case event, ok := <-source:
				if !ok {
					stream.close()
					return
				}
				stream.broadcast(event)
			}
		}
	}()
}

func (stream *Stream[T]) close() {
	stream.mu.Lock()
	for _, listener := range stream.listeners {
		close(listener)
	}
	stream.mu.Unlock()
}

func (stream *Stream[T]) broadcast(event T) {
	stream.mu.RLock()
	for _, listener := range stream.listeners {
		listener <- event
	}
	stream.mu.RUnlock()
}
