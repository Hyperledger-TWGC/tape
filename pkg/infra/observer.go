package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Observers struct {
	workers []*Observer
}

type Observer struct {
	Address string
	d       peer.Deliver_DeliverFilteredClient
	logger  *log.Logger
	cancel  context.CancelFunc
}

func CreateObservers(channel string, nodes []Node, crypto *Crypto, logger *log.Logger) (*Observers, error) {
	var workers []*Observer
	for _, node := range nodes {
		worker, err := CreateObserver(channel, node, crypto, logger)
		if err != nil {
			return nil, err
		}
		workers = append(workers, worker)
	}
	return &Observers{workers: workers}, nil
}

func (o *Observers) Start(N int, errorCh chan error, finishCh chan struct{}, now time.Time, blockCollector *BlockCollector) {
	for i := 0; i < len(o.workers); i++ {
		go o.workers[i].Start(N, errorCh, finishCh, now, blockCollector)
	}
}

func CreateObserver(channel string, node Node, crypto *Crypto, logger *log.Logger) (*Observer, error) {
	seek, err := CreateSignedDeliverNewestEnv(channel, crypto)
	if err != nil {
		return nil, err
	}

	ctx, cancel := TapeContext()
	deliverer, err := CreateDeliverFilteredClient(node, logger, ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	if err = deliverer.Send(seek); err != nil {
		cancel()
		return nil, err
	}

	// drain first response
	if _, err = deliverer.Recv(); err != nil {
		cancel()
		return nil, err
	}

	return &Observer{Address: node.Addr, d: deliverer, logger: logger, cancel: cancel}, nil
}

func (o *Observer) Start(N int, errorCh chan error, finishCh chan struct{}, now time.Time, blockCollector *BlockCollector) {
	defer close(finishCh)
	defer o.cancel()
	o.logger.Debugf("start observer")
	n := 0
	for n < N {
		r, err := o.d.Recv()
		if err != nil {
			errorCh <- err
		}

		if r == nil {
			errorCh <- errors.Errorf("received nil message, but expect a valid block instead. You could look into your peer logs for more info")
			return
		}

		fb := r.Type.(*peer.DeliverResponse_FilteredBlock)
		o.logger.Debugf("receivedTime %8.2fs\tBlock %6d\tTx %6d\t Address %s\n", time.Since(now).Seconds(), fb.FilteredBlock.Number, len(fb.FilteredBlock.FilteredTransactions), o.Address)

		if blockCollector.Commit(fb.FilteredBlock.Number) {
			// committed
			fmt.Printf("Time %8.2fs\tBlock %6d\tTx %6d\t \n", time.Since(now).Seconds(), fb.FilteredBlock.Number, len(fb.FilteredBlock.FilteredTransactions))
		}

		n = n + len(fb.FilteredBlock.FilteredTransactions)
	}
}
