package workers

import (
	"context"
	"runtime"
	"sync"
)

// IPool is the interface implemented by worker pools.
type IPool[T any] interface {
	Run()
	Queue() chan<- T
	Close()
	Wait()
	WorkerCount() int
	QueueSize() int
	OnError(ErrorHandler) IPool[T]
	OnRecovery(RecoveryHandler) IPool[T]
}

// Pool processes typed queue items from a channel with a fixed number of workers.
type Pool[T any] struct {
	ctx        context.Context
	callback   Callback[T]
	queue      chan T
	workers    int
	queueSize  int
	onError    ErrorHandler
	onRecovery RecoveryHandler
	runOnce    sync.Once
	closeOnce  sync.Once
	wg         sync.WaitGroup
}

// PoolParams contains configuration for a Pool.
type PoolParams[T any] struct {
	Callback  Callback[T]
	Workers   int
	QueueSize int
}

// DefaultPoolSize returns the default number of workers used by NewPool.
// It is runtime.NumCPU() - 1, with a minimum of 1.
func DefaultPoolSize() int {
	workers := runtime.NumCPU() - 1
	if workers < 1 {
		return 1
	}
	return workers
}

// NewPool creates a pool configured by params.
// Non-positive worker counts are replaced with DefaultPoolSize().
func NewPool[T any](ctx context.Context, params PoolParams[T]) *Pool[T] {
	if params.Workers <= 0 {
		params.Workers = DefaultPoolSize()
	}
	if params.QueueSize <= 0 {
		params.QueueSize = params.Workers
	}

	return &Pool[T]{
		ctx:       ctx,
		callback:  params.Callback,
		queue:     make(chan T, params.QueueSize),
		workers:   params.Workers,
		queueSize: params.QueueSize,
	}
}

// Run starts the pool worker goroutines. Safe to call multiple times; only the
// first call has effect.
func (pool *Pool[T]) Run() {
	pool.runOnce.Do(func() {
		pool.wg.Add(pool.workers)
		for i := 0; i < pool.workers; i++ {
			go pool.runWorker()
		}
	})
}

func (pool *Pool[T]) runWorker() {
	defer pool.wg.Done()
	defer pool.recovery()

	for {
		select {
		case payload, ok := <-pool.queue:
			if !ok {
				return
			}
			pool.execute(payload)
		case <-pool.ctx.Done():
			return
		}
	}
}

func (pool *Pool[T]) execute(payload T) {
	defer pool.recovery()
	if err := pool.callback(payload); err != nil && pool.onError != nil {
		pool.onError(err)
	}
}

func (pool *Pool[T]) recovery() {
	if pool.onRecovery == nil {
		return
	}

	if err := recover(); err != nil {
		pool.onRecovery(err)
	}
}

// Queue returns the channel used to send items to the pool.
func (pool *Pool[T]) Queue() chan<- T {
	return pool.queue
}

// Close closes the pool queue channel. Workers exit after queued items are
// processed or when ctx is cancelled.
func (pool *Pool[T]) Close() {
	pool.closeOnce.Do(func() {
		close(pool.queue)
	})
}

// Wait blocks until all worker goroutines exit.
func (pool *Pool[T]) Wait() {
	pool.wg.Wait()
}

// WorkerCount returns the configured number of worker goroutines.
func (pool *Pool[T]) WorkerCount() int {
	return pool.workers
}

// QueueSize returns the configured queue capacity.
func (pool *Pool[T]) QueueSize() int {
	return pool.queueSize
}

func (pool *Pool[T]) OnError(onError ErrorHandler) IPool[T] {
	pool.onError = onError
	return pool
}

func (pool *Pool[T]) OnRecovery(onRecovery RecoveryHandler) IPool[T] {
	pool.onRecovery = onRecovery
	return pool
}
