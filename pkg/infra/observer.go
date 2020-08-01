package infra

import (
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger/fabric-protos-go/peer"
	log "github.com/sirupsen/logrus"
)

type BlockCollection struct {
	Collection map[uint64]int
	lock       sync.Mutex
}

type Observers struct {
	workers []*Observer
	logger  *log.Logger
}

type Observer struct {
	Address   string
	d         peer.Deliver_DeliverFilteredClient
	logger    *log.Logger
	signal    chan error
	countDown int
}

func CreateObservers(channel string, nodes []Node, countDown int, crypto *Crypto, logger *log.Logger) *Observers {
	var workers []*Observer
	for _, node := range nodes {
		workers = append(workers, CreateObserver(channel, node, countDown, crypto, logger))
	}
	return &Observers{workers: workers, logger: logger}
}

func (o *Observers) Start(N int, now time.Time, blockcollection *BlockCollection) {
	for i := 0; i < len(o.workers); i++ {
		go o.workers[i].Start(N, now, blockcollection)
	}
}

func (o *Observers) Wait() {
	for i := 0; i < len(o.workers); i++ {
		o.workers[i].Wait()
	}
}

func CreateObserver(channel string, node Node, countDown int, crypto *Crypto, logger *log.Logger) *Observer {
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

	return &Observer{Address: node.Addr, d: deliverer, countDown: countDown, signal: make(chan error, 10), logger: logger}
}

func (o *Observer) Start(N int, now time.Time, blockcollection *BlockCollection) {
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

		fb := r.Type.(*peer.DeliverResponse_FilteredBlock)
		// lock
		blockcollection.lock.Lock()
		// if meet
		if i, ok := blockcollection.Collection[fb.FilteredBlock.Number]; ok {
			if i > 0 {
				// if countDown
				i := i - 1
				if i == 0 {
					// if countDown zero
					fmt.Printf("Time %8.2fs\tBlock %6d\tTx %6d\t \n", time.Since(now).Seconds(), fb.FilteredBlock.Number, len(fb.FilteredBlock.FilteredTransactions))
				} else {
					o.logger.Debugf("Not meet numbers recevied block, skip, %d peers need receive the block. \n", i)
				}
				blockcollection.Collection[fb.FilteredBlock.Number] = i
			}
		} else {
			// otherwise new
			blockcollection.Collection[fb.FilteredBlock.Number] = o.countDown - 1
			if o.countDown == 1 {
				fmt.Printf("Time %8.2fs\tBlock %6d\tTx %6d\t \n", time.Since(now).Seconds(), fb.FilteredBlock.Number, len(fb.FilteredBlock.FilteredTransactions))
			}
		}
		// release lock
		blockcollection.lock.Unlock()
		n = n + len(fb.FilteredBlock.FilteredTransactions)
		o.logger.Debugf("Time %8.2fs\tBlock %6d\tTx %6d\t Address %s\n", time.Since(now).Seconds(), fb.FilteredBlock.Number, len(fb.FilteredBlock.FilteredTransactions), o.Address)
	}
}

func (o *Observer) Wait() {
	for err := range o.signal {
		if err != nil {
			o.logger.Errorf("Observed error: %s\n", err)
		}
	}
}
