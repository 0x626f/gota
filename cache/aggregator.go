package cache

import (
	"time"

	"golang.org/x/sync/singleflight"
)

// RecoveryHandler handles a panic recovered while an Aggregator call is running.
//
// The shape argument is the key passed to Call. The subject argument is the
// recovered panic value.
type RecoveryHandler func(shape string, subject any)

// Aggregator combines a string-keyed cache with singleflight duplicate
// suppression.
//
// Concurrent calls for the same shape share one in-flight action. Calls for
// different shapes may run independently. When a cache is configured, successful
// results are stored in the cache and later calls for the same shape return the
// cached value without running action.
type Aggregator[D any] struct {
	syncer singleflight.Group
	cache  Cache[D, string]

	onRecover RecoveryHandler
}

// NewAggregator creates an Aggregator.
//
// The cache argument is optional. When no cache is provided, the Aggregator only
// suppresses duplicate in-flight calls. When a cache is provided, Call checks it
// before running action and stores successful action results in it.
func NewAggregator[D any](cache ...Cache[D, string]) *Aggregator[D] {
	agg := &Aggregator[D]{}

	if len(cache) > 0 && cache[0] != nil {
		agg.cache = cache[0]
	}

	return agg
}

// OnRecovery registers a panic recovery handler and returns the Aggregator.
//
// If action panics and a handler is registered, Call recovers the panic after
// singleflight has cleaned up the in-flight call, passes the recovered value to
// the handler, and returns the zero value of D with a nil error. Without a
// handler, the panic is propagated.
func (aggregator *Aggregator[D]) OnRecovery(handler RecoveryHandler) *Aggregator[D] {
	aggregator.onRecover = handler
	return aggregator
}

// Call returns a cached or loaded value for shape.
//
// If the configured cache already contains shape, the cached value is returned.
// Otherwise action is run through singleflight so concurrent calls for the same
// shape share one result. Successful action results are cached; errors are
// returned to callers and are not cached. Optional ttl is passed to the cache
// when storing a successful result.
func (aggregator *Aggregator[D]) Call(shape string, action func() (D, error), ttl ...time.Duration) (D, error) {
	defer func() {
		if r := recover(); r != nil {
			if aggregator.onRecover == nil {
				panic(r)
			}
			aggregator.onRecover(shape, r)
		}
	}()

	caller := func() (any, error) {
		if aggregator.cache != nil {
			if result, exists := aggregator.cache.Get(shape); exists {
				return result, nil
			}
		}

		result, err := action()

		if err == nil && aggregator.cache != nil {
			if len(ttl) > 0 {
				aggregator.cache.Set(shape, result, ttl[0])
			} else {
				aggregator.cache.Set(shape, result)
			}
		}
		return result, err
	}

	result, err, _ := aggregator.syncer.Do(shape, caller)
	return result.(D), err
}
