package infra

import (
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

// BlockCollector keeps track of committed blocks on multiple peers.
// This is used when a block is considered confirmed only when committed
// on a certain number of peers within network.
type BlockCollector struct {
	sync.Mutex
	thresholdP, totalP int
	totalTx            int
	registry           map[uint64]int
}

// NewBlockCollector creates a BlockCollector
func NewBlockCollector(threshold int, total int) (*BlockCollector, error) {
	registry := make(map[uint64]int)
	if threshold > total {
		return nil, errors.Errorf("threshold [%d] must be less than or equal to total [%d]", threshold, total)
	}
	return &BlockCollector{
		thresholdP: threshold,
		totalP:     total,
		registry:   registry,
	}, nil
}

func (bc *BlockCollector) Start(
	blockCh <-chan *peer.FilteredBlock,
	finishCh chan struct{},
	totalTx int,
	now time.Time,
	printResult bool, // controls whether to print block commit message. Tests set this to false to avoid polluting stdout.
) {
	// TODO block collector should be able to detect repeated block, and exclude it from total tx counting.
	for block := range blockCh {
		cnt := bc.registry[block.Number] // cnt is default to 0 when key does not exist
		cnt++

		// newly committed block just hits threshold
		if cnt == bc.thresholdP {
			if printResult {
				fmt.Printf("Time %8.2fs\tBlock %6d\tTx %6d\t \n", time.Since(now).Seconds(), block.Number, len(block.FilteredTransactions))
			}

			bc.totalTx += len(block.FilteredTransactions)
			if bc.totalTx >= totalTx {
				close(finishCh)
			}
		}

		if cnt == bc.totalP {
			// committed on all peers, remove from registry
			delete(bc.registry, block.Number)
		} else {
			// upsert back to registry
			bc.registry[block.Number] = cnt
		}
	}
}

// Deprecated
//
// Commit commits a block to collector. It returns true iff the number of peers on which
// this block has been committed has satisfied thresholdP.
func (bc *BlockCollector) Commit(block uint64) (committed bool) {
	bc.Lock()
	defer bc.Unlock()

	cnt := bc.registry[block] // cnt is default to 0 when key does not exist
	cnt++

	// newly committed block just hits threshold
	if cnt == bc.thresholdP {
		committed = true
	}

	if cnt == bc.totalP {
		// committed on all peers, remove from registry
		delete(bc.registry, block)
	} else {
		// upsert back to registry
		bc.registry[block] = cnt
	}

	return
}
