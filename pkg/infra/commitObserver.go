package infra

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric-protos-go/orderer"
	log "github.com/sirupsen/logrus"
)

type CommitObserver struct {
	d      orderer.AtomicBroadcast_DeliverClient
	logger *log.Logger
	signal chan error
}

func CreateCommitObserver(channel string, node Node, crypto *Crypto, logger *log.Logger) *CommitObserver {
	if len(node.Addr) == 0 {
		return nil
	}
	deliverer, err := CreateDeliverClient(node)
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

	return &CommitObserver{d: deliverer, signal: make(chan error, 10), logger: logger}
}

func (o *CommitObserver) Start(N int, now time.Time) {
	defer close(o.signal)
	o.logger.Debugf("start observer")
	n := 0
	for n < N {
		r, err := o.d.Recv()
		if err != nil {
			o.signal <- err
		}
		if r == nil {
			panic("Received nil message, but expect a valid block instead. You could look into your peer logs for more info")
		}
		tx := len(r.GetBlock().Data.Data)
		n += tx
		fmt.Printf("Time %8.2fs\tBlock %6d\t Tx %6d\n", time.Since(now).Seconds(), n, tx)
	}
}

func (o *CommitObserver) Wait() {
	for err := range o.signal {
		if err != nil {
			o.logger.Errorf("Observed error: %s\n", err)
		}
	}
}
