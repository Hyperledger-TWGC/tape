package infra

import (
	"sync"

	"github.com/pkg/errors"
)

type BlockCollector struct {
	sync.Mutex
	threshold int
	total     int
	registry  map[uint64]int
}

func NewBlockCollector(threshold int, total int) (*BlockCollector, error) {
	registry := make(map[uint64]int)
	if threshold > total {
		return nil, errors.Errorf("commitThreshold should not bigger than committers, please check your config")
	}
	return &BlockCollector{
		threshold: threshold,
		total:     total,
		registry:  registry,
	}, nil
}

func (bc *BlockCollector) Commit(block uint64) (committed bool) {
	bc.Lock()
	defer bc.Unlock()
	committed = false
	// read map
	committedCount, ok := bc.registry[block]
	if !ok {
		committedCount = 0
	}
	committedCount = committedCount + 1
	// match the threshold
	if committedCount == bc.threshold {
		committed = true
	}
	// all peers returned avoid save
	if committedCount == bc.total {
		delete(bc.registry, block)
	} else {
		bc.registry[block] = committedCount
	}
	return committed
}
