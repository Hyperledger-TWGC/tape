package observer

import (
	"fmt"
	"math"
	"tape/internal/fabric/protoutil"
	"tape/pkg/infra/basic"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	log "github.com/sirupsen/logrus"
)

type CommitObserver struct {
	d        orderer.AtomicBroadcast_DeliverClient
	n        int
	logger   *log.Logger
	Now      time.Time
	errorCh  chan error
	finishCh chan struct{}
}

func CreateCommitObserver(
	channel string,
	node basic.Node,
	crypto *basic.CryptoImpl,
	logger *log.Logger,
	n int,
	errorCh chan error,
	finishCh chan struct{}) (*CommitObserver, error) {
	if len(node.Addr) == 0 {
		return nil, nil
	}
	deliverer, err := basic.CreateDeliverClient(node)
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

	return &CommitObserver{d: deliverer,
		n:        n,
		logger:   logger,
		errorCh:  errorCh,
		finishCh: finishCh}, nil
}

func (o *CommitObserver) Start() {
	o.Now = time.Now()
	o.logger.Debugf("start observer")
	n := 0
	for n < o.n {
		r, err := o.d.Recv()
		if err != nil {
			o.errorCh <- err
		}
		if r == nil {
			panic("Received nil message, but expect a valid block instead. You could look into your peer logs for more info")
		}
		tx := len(r.GetBlock().Data.Data)
		n += tx
		fmt.Printf("Time %8.2fs\tBlock %6d\t Tx %6d\n", time.Since(o.Now).Seconds(), n, tx)
	}
	close(o.finishCh)
}

func (o *CommitObserver) GetTime() time.Time {
	return o.Now
}

func CreateSignedDeliverNewestEnv(ch string, signer *basic.CryptoImpl) (*common.Envelope, error) {
	start := &orderer.SeekPosition{
		Type: &orderer.SeekPosition_Newest{
			Newest: &orderer.SeekNewest{},
		},
	}

	stop := &orderer.SeekPosition{
		Type: &orderer.SeekPosition_Specified{
			Specified: &orderer.SeekSpecified{
				Number: math.MaxUint64,
			},
		},
	}

	seekInfo := &orderer.SeekInfo{
		Start:    start,
		Stop:     stop,
		Behavior: orderer.SeekInfo_BLOCK_UNTIL_READY,
	}

	return protoutil.CreateSignedEnvelope(
		common.HeaderType_DELIVER_SEEK_INFO,
		ch,
		signer,
		seekInfo,
		0,
		0,
	)
}
