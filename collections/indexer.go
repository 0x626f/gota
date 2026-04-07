package collections

import "sync"

type Indexer struct {
	index  int
	staged []int

	mu sync.Mutex
}

func NewIndexer(preSize ...int) *Indexer {
	indexer := &Indexer{}

	if len(preSize) != 0 && preSize[0] > 0 {
		indexer.staged = make([]int, 0, preSize[0])
	}

	return indexer
}

func (indexer *Indexer) Next() (index int) {
	indexer.mu.Lock()
	defer indexer.mu.Unlock()

	if len(indexer.staged) > 0 {
		indexer.staged, index = indexer.staged[:len(indexer.staged)-1], indexer.staged[len(indexer.staged)-1]
	} else {
		index = indexer.index
		indexer.index++
	}
	return
}

func (indexer *Indexer) Release(index int) {
	indexer.mu.Lock()
	defer indexer.mu.Unlock()

	indexer.staged = append(indexer.staged, index)
}
