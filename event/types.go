// Package event provides primitives for event-driven programming: a type-safe
// Router that dispatches events to registered handlers (with optional async
// dispatch and error handling), and a generic Stream that broadcasts values
// from a single source channel to multiple concurrent listeners.
package event

// ID is the unique identifier for an event type.
type ID = string

// Event represents a dispatchable message with a type identifier and payload.
type Event interface {
	// Id returns the event's type identifier, used to route it to the correct handler.
	Id() ID
	// Data returns the event payload.
	Data() any
}

// Handler is a function that processes an event's payload.
// It returns an error if processing fails.
type Handler func(any) error

// ErrorHandler is a function called when an error occurs during event routing
// or handler execution. It receives the original event and the error.
type ErrorHandler func(Event, error)
