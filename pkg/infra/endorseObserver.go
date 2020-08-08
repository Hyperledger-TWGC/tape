package infra

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type EndorseObserver struct {
	p      chan *Elements
	n      int
	lock   sync.Mutex
	logger *log.Logger
	signal bool
}

func CreateEndorseObserver(processed chan *Elements, logger *log.Logger) *EndorseObserver {
	return &EndorseObserver{p: processed, logger: logger}
}

func (o *EndorseObserver) Start(N int, now time.Time) {
	o.logger.Debugf("start observer")
	o.lock.Lock()
	for o.n < N {
		select {
		case e := <-o.p:
			o.n = o.n + 1
			o.logger.Debugln(e.Proposal.GetHeader())
			fmt.Printf("Time %8.2fs\tTx %6d Processed\n", time.Since(now).Seconds(), o.n)
		}
	}
	o.lock.Unlock()
	o.signal = true
}

func (o *EndorseObserver) Wait() {
	for !o.signal {
	}
}
