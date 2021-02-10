package infra

import (
	"context"
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
	ctx context.Context,
	blockCh <-chan *peer.FilteredBlock,
	finishCh chan struct{},
	totalTx int,
	now time.Time,
	printResult bool, // controls whether to print block commit message. Tests set this to false to avoid polluting stdout.
) {
	// TODO block collector should be able to detect repeated block, and exclude it from total tx counting.
	for {
		select {
		case block := <-blockCh:
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
		case <-ctx.Done():
			return
		}
	}
}

// Deprecated
//
// Commit commits a block to collector. It returns true iff the number of peers on which
// this block has been committed has satisfied thresholdP.
func (bc *BlockCollector) Commit(block *peer.DeliverResponse_FilteredBlock, finishCh chan struct{}, now time.Time) (committed bool) {
	bc.Lock()
	defer bc.Unlock()

	cnt := bc.registry[block.FilteredBlock.Number] // cnt is default to 0 when key does not exist
	cnt++

	// newly committed block just hits threshold
	if cnt == bc.thresholdP {
		committed = true
		duration := time.Since(now)
		bc.totalTx += len(block.FilteredBlock.FilteredTransactions)
		fmt.Printf("tx: %d, duration: %+v, tps: %f\n", bc.totalTx, duration, float64(bc.totalTx)/duration.Seconds())
	}

	if cnt == bc.totalP {
		// committed on all peers, remove from registry
		delete(bc.registry, block.FilteredBlock.Number)
	} else {
		// upsert back to registry
		bc.registry[block.FilteredBlock.Number] = cnt
	}

	return
}
