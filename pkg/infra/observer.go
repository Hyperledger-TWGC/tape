package infra

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric-protos-go/peer"
	log "github.com/sirupsen/logrus"
)

type Observer struct {
	d      peer.Deliver_DeliverFilteredClient
	logger *log.Logger
	signal chan error
}

func CreateObserver(channel string, node Node, crypto *Crypto, logger *log.Logger) *Observer {
	TLSCACert, err := GetTLSCACerts(node.TLSCACert)
	if err != nil {
		panic(err)
	}
	deliverer, err := CreateDeliverFilteredClient(node.Addr, TLSCACert)
	if err != nil {
		panic(err)
	}

	seek, err := CreateSignedDeliverNewestEnv(channel, crypto)
	if err != nil {
		panic(err)
	}

	if err = deliverer.Send(seek); err != nil {
		panic(err)
	}

	// drain first response
	if _, err = deliverer.Recv(); err != nil {
		panic(err)
	}

	return &Observer{d: deliverer, signal: make(chan error, 10), logger: logger}
}

func (o *Observer) Start(N int, now time.Time) {
	defer close(o.signal)
	o.logger.Debugf("start observer")
	n := 0
	for n < N {
		r, err := o.d.Recv()
		if err != nil {
			o.signal <- err
		}

		fb := r.Type.(*peer.DeliverResponse_FilteredBlock)
		n = n + len(fb.FilteredBlock.FilteredTransactions)
		fmt.Printf("Time %8.2fs\tBlock %6d\tTx %6d\n", time.Since(now).Seconds(), fb.FilteredBlock.Number, len(fb.FilteredBlock.FilteredTransactions))
	}
}

func (o *Observer) Wait() {
	for err := range o.signal {
		if err != nil {
			o.logger.Errorf("Observed error: %s\n", err)
		}
	}
}
