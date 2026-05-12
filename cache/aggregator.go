package cache

import (
	"time"

	"golang.org/x/sync/singleflight"
)

type RecoveryHandler func(shape string, subject any)

type Aggregator[D any] struct {
	syncer singleflight.Group
	cache  Cache[D, string]

	onRecover RecoveryHandler
}

func NewAggregator[D any](cache ...Cache[D, string]) *Aggregator[D] {
	agg := &Aggregator[D]{}

	if len(cache) > 0 && cache[0] != nil {
		agg.cache = cache[0]
	}

	return agg
}

func (aggregator *Aggregator[D]) OnRecovery(handler RecoveryHandler) *Aggregator[D] {
	aggregator.onRecover = handler
	return aggregator
}

func (aggregator *Aggregator[D]) Call(shape string, action func() (D, error), ttl ...time.Duration) (D, error) {
	defer func() {
		if r := recover(); r != nil {
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
