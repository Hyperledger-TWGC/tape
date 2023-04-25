package observer

import (
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger-twgc/tape/pkg/infra/basic"

	log "github.com/sirupsen/logrus"
)

type EndorseObserver struct {
	Envs     chan *basic.TracingEnvelope
	n        int
	logger   *log.Logger
	Now      time.Time
	once     *sync.Once
	finishCh chan struct{}
}

func CreateEndorseObserver(Envs chan *basic.TracingEnvelope, N int, finishCh chan struct{}, once *sync.Once, logger *log.Logger) *EndorseObserver {
	return &EndorseObserver{
		Envs:     Envs,
		n:        N,
		logger:   logger,
		finishCh: finishCh,
		once:     once,
	}
}

func (o *EndorseObserver) Start() {
	o.Now = time.Now()
	o.logger.Debugf("start observer for endorsement")
	i := 0
	for {
		select {
		case e := <-o.Envs:
			tapeSpan := basic.GetGlobalSpan()
			tapeSpan.FinishWithMap(e.TxId, "", basic.TRANSCATIONSTART)
			i++
			fmt.Printf("Time %8.2fs\tTx %6d Processed\n", time.Since(o.Now).Seconds(), i)
			if o.n > 0 {
				if o.n == i {
					// consider with multiple threads need close this channel, need a once here to avoid channel been closed in multiple times
					o.once.Do(func() {
						close(o.finishCh)
					})
					return
				}
			}
		}
	}
}

func (o *EndorseObserver) GetTime() time.Time {
	return o.Now
}
