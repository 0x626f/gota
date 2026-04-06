// Package workers provides utilities for running background tasks. It includes
// helpers to schedule work on a recurring interval, trigger work from a
// channel signal, and execute fallible operations with automatic retries or
// execution-time measurement.
package workers

import (
	"context"
	"sync"
	"time"
)

// IWorker is the interface implemented by all workers.
type IWorker interface {
	Run()
	OnError(ErrorHandler)
	OnRecovery(RecoveryHandler)
}

// Worker is the base implementation of IWorker. Use the New* constructors to
// obtain a configured instance rather than creating one directly.
type Worker struct {
	ctx        context.Context
	runner     Runner
	onError    ErrorHandler
	onRecovery RecoveryHandler
	once       sync.Once
}

// NewWorker creates a bare Worker with no runner attached.
func NewWorker(ctx context.Context, onError ErrorHandler, onRecovery RecoveryHandler) *Worker {
	return &Worker{
		ctx:        ctx,
		onError:    onError,
		onRecovery: onRecovery,
	}
}

func executeTask(worker *Worker, task Task) {
	defer worker.recovery()
	if err := task(); err != nil && worker.onError != nil {
		worker.onError(err)
	}
}

func executeCallback[T any](worker *Worker, callback Callback[T], payload T) {
	defer worker.recovery()
	if err := callback(payload); err != nil && worker.onError != nil {
		worker.onError(err)
	}
}

// Run starts the background goroutine. Safe to call multiple times; only the
// first call has effect.
func (worker *Worker) Run() {
	worker.once.Do(func() {
		if worker.runner != nil {
			go worker.runner()
		}
	})
}

func (worker *Worker) recovery() {
	if worker.onRecovery == nil {
		return
	}

	if err := recover(); err != nil {
		worker.onRecovery(err)
	}
}

func (worker *Worker) OnError(onError ErrorHandler) {
	worker.onError = onError
}

func (worker *Worker) OnRecovery(onRecovery RecoveryHandler) {
	worker.onRecovery = onRecovery
}

// NewWorkerOnTicker creates a worker that executes task immediately on Run(),
// then again on every tick of delay. Stops when ctx is cancelled.
func NewWorkerOnTicker(ctx context.Context, task Task, delay time.Duration) IWorker {
	worker := NewWorker(ctx, nil, nil)
	worker.runner = func() {
		defer worker.recovery()

		ticker := time.NewTicker(delay)
		defer ticker.Stop()
		executeTask(worker, task)
		for {
			select {
			case <-ticker.C:
				go executeTask(worker, task)
			case <-ctx.Done():
				return
			}
		}
	}
	return worker
}

// NewWorkerOnSignal creates a worker that executes task
// each time a value is received on signal.
// Stops when ctx is cancelled or signal is closed.
func NewWorkerOnSignal(ctx context.Context, task Task, signal <-chan struct{}) IWorker {
	worker := NewWorker(ctx, nil, nil)
	worker.runner = func() {
		defer worker.recovery()

		for {
			select {
			case _, ok := <-signal:
				if ok {
					go executeTask(worker, task)
				} else {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}
	return worker
}

// NewWorkerOnEvent creates a worker that invokes callback with each value
// received on signal.
// Stops when ctx is cancelled or signal is closed.
func NewWorkerOnEvent[T any](ctx context.Context, callback Callback[T], signal <-chan T) IWorker {
	worker := NewWorker(ctx, nil, nil)
	worker.runner = func() {
		defer worker.recovery()

		for {
			select {
			case payload, ok := <-signal:
				if ok {
					go executeCallback(worker, callback, payload)
				} else {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}
	return worker
}
