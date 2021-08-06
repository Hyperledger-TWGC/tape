package observer

import (
	"fmt"
	"tape/pkg/infra/basic"
	"time"

	log "github.com/sirupsen/logrus"
)

type EndorseObserver struct {
	p        chan *basic.Elements
	n        int
	logger   *log.Logger
	Now      time.Time
	finishCh chan struct{}
}

func CreateEndorseObserver(processed chan *basic.Elements, N int, finishCh chan struct{}, logger *log.Logger) *EndorseObserver {
	return &EndorseObserver{p: processed, n: N, logger: logger, finishCh: finishCh}
}

func (o *EndorseObserver) Start() {
	o.Now = time.Now()
	o.logger.Debugf("start observer")
	i := 0
	for o.n >= i {
		select {
		case e := <-o.p:
			o.logger.Debugln(e)
			fmt.Printf("Time %8.2fs\tTx %6d Processed\n", time.Since(o.Now).Seconds(), i)
			i++
		}
	}
	close(o.finishCh)
}

func (o *EndorseObserver) GetTime() time.Time {
	return o.Now
}
