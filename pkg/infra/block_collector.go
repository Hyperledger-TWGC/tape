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
	PeerIdx int // source peer's number
}

type collectorBlock struct {
	collector int
	block     *peer.DeliverResponse_FilteredBlock
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
	successRateBlockCh chan<- *AddressedBlock,
	finishCh chan struct{},
	totalTx int,
	now time.Time,
	printResult bool, // controls whether to print block commit message. Tests set this to false to avoid polluting stdout.
) {
	for {
		select {
		case block := <-blockCh:
			bc.commit(block, successRateBlockCh, finishCh, totalTx, now, printResult)
		case <-ctx.Done():
			return
		}
	}
}

// TODO This function contains too many functions and needs further optimization
// commit commits a block to collector.
// If the number of peers on which this block has been committed has satisfied thresholdP,
// adds the number to the totalTx.
func (bc *BlockCollector) commit(block *AddressedBlock, successRatioBlockCh chan<- *AddressedBlock, finishCh chan struct{}, totalTx int, now time.Time, printResult bool) {
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
	if bitMap.Has(block.PeerIdx) {
		return
	}

	bitMap.Set(block.PeerIdx)
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
	successRatioBlockCh <- block
}

// CalSuccessRate calculate the success rate of the txs
// First calculate the transaction success rate accepted by each peer,
// and then count all the transaction success rates
func CalSuccessRate(collectNum int, totalTx int, blocks chan *AddressedBlock) {
	var totalTxs, totalSuccessTxs int
	allTxs := collectNum * totalTx
	successTxs := make([]int, collectNum)

	for {
		select {
		case block := <-blocks:
			for _, tx := range block.FilteredTransactions {
				if tx.TxValidationCode == peer.TxValidationCode_VALID {
					successTxs[block.PeerIdx]++
				}
			}
			totalTxs += len(block.FilteredTransactions)
		}

		if totalTxs >= allTxs {
			break
		}
	}

	fmt.Println("The txs' success rate is as followers:")
	for i := 0; i < collectNum; i++ {
		totalSuccessTxs += successTxs[i]
		fmt.Printf("peer %d received %d txs, containing %d successful txs, and the success rate is %.2f%%\n", i, totalTx, successTxs[i], float64(successTxs[i])/float64(totalTx)*100)
	}
	fmt.Printf("All peer received %d txs, containing %d successful txs, and the success rate is %.2f%%\n", totalTxs, totalSuccessTxs, float64(totalSuccessTxs)/float64(totalTxs)*100)
}
