package observer

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
	log "github.com/sirupsen/logrus"
)

type EndorseObserver struct {
	Envs     chan *common.Envelope
	n        int
	logger   *log.Logger
	Now      time.Time
	finishCh chan struct{}
}

func CreateEndorseObserver(Envs chan *common.Envelope, N int, finishCh chan struct{}, logger *log.Logger) *EndorseObserver {
	return &EndorseObserver{Envs: Envs, n: N, logger: logger, finishCh: finishCh}
}

func (o *EndorseObserver) Start() {
	o.Now = time.Now()
	o.logger.Debugf("start observer")
	i := 0
	for o.n > i {
		select {
		case <-o.Envs:
			//o.logger.Debugln(e)
			fmt.Printf("Time %8.2fs\tTx %6d Processed\n", time.Since(o.Now).Seconds(), i)
			i++
		}
	}
	close(o.finishCh)
}

func (o *EndorseObserver) GetTime() time.Time {
	return o.Now
}
