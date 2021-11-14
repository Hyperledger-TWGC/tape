package observer

import (
	"fmt"
	"tape/pkg/infra/basic"
	"time"

	log "github.com/sirupsen/logrus"
)

type EndorseObserver struct {
	Envs     chan *basic.TracingEnvelope
	n        int
	logger   *log.Logger
	Now      time.Time
	finishCh chan struct{}
}

func CreateEndorseObserver(Envs chan *basic.TracingEnvelope, N int, finishCh chan struct{}, logger *log.Logger) *EndorseObserver {
	return &EndorseObserver{Envs: Envs, n: N, logger: logger, finishCh: finishCh}
}

func (o *EndorseObserver) Start() {
	o.Now = time.Now()
	o.logger.Debugf("start observer for endorsement")
	i := 0
	for {
		select {
		case <-o.Envs:
			//o.logger.Debugln(e)
			i++
			fmt.Printf("Time %8.2fs\tTx %6d Processed\n", time.Since(o.Now).Seconds(), i)
			if o.n > 0 {
				if o.n == i {
					close(o.finishCh)
					return
				}
			}
		}

	}
}

func (o *EndorseObserver) GetTime() time.Time {
	return o.Now
}
