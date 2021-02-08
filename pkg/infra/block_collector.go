package infra

import (
	"math/big"
	"sync"

	"github.com/pkg/errors"
)

// BlockCollector keeps track of committed blocks on multiple peers.
// This is used when a block is considered confirmed only when committed
// on a certain number of peers within network.
type BlockCollector struct {
	sync.Mutex
	threshold    int
	total        int
	registry     map[uint64]int
	confirmedNum *big.Int
}

// NewBlockCollector creates a BlockCollector
func NewBlockCollector(threshold int, total int) (*BlockCollector, error) {
	registry := make(map[uint64]int)
	if threshold > total {
		return nil, errors.Errorf("threshold [%d] must be less than or equal to total [%d]", threshold, total)
	}
	return &BlockCollector{
		threshold:    threshold,
		total:        total,
		registry:     registry,
		confirmedNum: new(big.Int).SetInt64(0),
	}, nil
}

// Commit commits a block to collector. It returns true iff the number of peers on which
// this block has been committed has satisfied threshold.
func (bc *BlockCollector) Commit(block uint64) (committed bool) {
	bc.Lock()
	defer bc.Unlock()

	cnt := bc.registry[block] // cnt is default to 0 when key does not exist
	cnt++

	// newly committed block just hits threshold
	if cnt == bc.threshold {
		committed = true
		bc.confirmedNum.Add(bc.confirmedNum, one)
	}

	if cnt == bc.total {
		// committed on all peers, remove from registry
		delete(bc.registry, block)
	} else {
		// upsert back to registry
		bc.registry[block] = cnt
	}

	return
}

func (bc *BlockCollector) GetConfirmNum() int64 {
	bc.Lock()
	defer bc.Unlock()
	return bc.confirmedNum.Int64()
}
