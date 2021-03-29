package infra

import (
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

func (o *Observers) Start(errorCh chan error, blockCh chan<- *peer.FilteredBlock, now time.Time) {
	for i := 0; i < len(o.workers); i++ {
		go o.workers[i].Start(errorCh, blockCh, now)
	}
}

func CreateObserver(channel string, node Node, crypto *Crypto, logger *log.Logger) (*Observer, error) {
	deliverer, err := CreateDeliverFilteredClient(node, logger)
	if err != nil {
		return nil, err
	}

	seek, err := CreateSignedDeliverNewestEnv(channel, crypto)
	if err != nil {
		return nil, err
	}

	if err = deliverer.Send(seek); err != nil {
		return nil, err
	}

	// drain first response
	if _, err = deliverer.Recv(); err != nil {
		return nil, err
	}

	return &Observer{Address: node.Addr, d: deliverer, logger: logger}, nil
}

func (o *Observer) Start(errorCh chan error, blockCh chan<- *peer.FilteredBlock, now time.Time) {
	o.logger.Debugf("start observer for orderer %s", o.Address)
	n := 0
	for {
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

		// TODO use proper context to create deliver client so it could be cancelled (now it's context.Background).
		blockCh <- fb.FilteredBlock

		n = n + len(fb.FilteredBlock.FilteredTransactions)
	}
}
