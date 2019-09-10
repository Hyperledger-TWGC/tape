package infra

import (
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger/fabric/protos/peer"
)

type Observer struct {
	d []peer.Deliver_DeliverFilteredClient

	signal chan error
	wg     *sync.WaitGroup
}

func CreateObserver(addr string, channels []string, crypto *Crypto) *Observer {
	observer := &Observer{signal: make(chan error, 10), wg: new(sync.WaitGroup)}
	observer.d = make([]peer.Deliver_DeliverFilteredClient, len(channels))
	for i := 0; i < len(channels); i++ {
		deliverer, err := CreateDeliverFilteredClient(addr, crypto.TLSCACerts)
		if err != nil {
			panic(err)
		}
		observer.d[i] = deliverer
	}

	seeks, err := CreateSignedDeliverNewestEnv(channels, crypto)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(seeks); i++ {
		if err = observer.d[i].Send(seeks[i]); err != nil {
			panic(err)
		}
	}

	// drain first response
	for i := 0; i < len(channels); i++ {
		if _, err := observer.d[i].Recv(); err != nil {
			panic(err)
		}
	}

	return observer
}

func (o *Observer) Start(N int, now time.Time, channels int) {
	o.wg.Add(channels)
	for i := 0; i < channels; i++ {
		go o.ReceiveBlock(i, now, N)
	}

}

func (o *Observer) Wait() {
	o.wg.Wait()
}

func (o *Observer) ReceiveBlock(index int, now time.Time, N int) {
	n := 0
	for n < N {
		r, err := o.d[index].Recv()
		if err != nil {
			fmt.Printf("Observed error: %s\n", err.Error())
		}

		fb := r.Type.(*peer.DeliverResponse_FilteredBlock)
		n = n + len(fb.FilteredBlock.FilteredTransactions)
		fmt.Printf("Time %v\tChannel %s\tBlock %d\tTx %d\n", time.Since(now), fb.FilteredBlock.ChannelId, fb.FilteredBlock.Number, len(fb.FilteredBlock.FilteredTransactions))
	}
	o.wg.Done()
}
