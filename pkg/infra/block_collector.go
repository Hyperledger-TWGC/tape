package infra

import (
	"context"
	"fmt"
	"sync"
	"tape/pkg/infra/bitmap"
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
	registry           map[uint64]*bitmap.BitMap
}

// AddressedBlock describe the source of block
type AddressedBlock struct {
	*peer.FilteredBlock
	Address int // source peer's number
}

// NewBlockCollector creates a BlockCollector
func NewBlockCollector(threshold int, total int) (*BlockCollector, error) {
	registry := make(map[uint64]*bitmap.BitMap)
	if threshold <= 0 || total <= 0 {
		return nil, errors.New("threshold and total must be greater than zero")
	}
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
	blockCh <-chan *AddressedBlock,
	finishCh chan struct{},
	totalTx int,
	now time.Time,
	printResult bool, // controls whether to print block commit message. Tests set this to false to avoid polluting stdout.
) {
	for {
		select {
		case block := <-blockCh:
			bc.commit(block, finishCh, totalTx, now, printResult)
		case <-ctx.Done():
			return
		}
	}
}

// TODO This function contains too many functions and needs further optimization
// commit commits a block to collector.
// If the number of peers on which this block has been committed has satisfied thresholdP,
// adds the number to the totalTx.
func (bc *BlockCollector) commit(block *AddressedBlock, finishCh chan struct{}, totalTx int, now time.Time, printResult bool) {
	bitMap, ok := bc.registry[block.Number]
	if !ok {
		// The block with Number is received for the first time
		b, err := bitmap.NewBitMap(bc.totalP)
		if err != nil {
			panic("Can not make new bitmap for BlockCollector" + err.Error())
		}
		bc.registry[block.Number] = &b
		bitMap = &b
	}
	// When the block from Address has been received before, return directly.
	if bitMap.Has(block.Address) {
		return
	}

	bitMap.Set(block.Address)
	cnt := bitMap.Count()

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

	// TODO issue176
	if cnt == bc.totalP {
		// committed on all peers, remove from registry
		delete(bc.registry, block.Number)
	}
}
