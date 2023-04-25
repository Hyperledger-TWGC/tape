package observer

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/hyperledger-twgc/tape/internal/fabric/protoutil"
	"github.com/hyperledger-twgc/tape/pkg/infra"
	"github.com/hyperledger-twgc/tape/pkg/infra/basic"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	log "github.com/sirupsen/logrus"
)

type CommitObserver struct {
	d          orderer.AtomicBroadcast_DeliverClient
	n          int
	logger     *log.Logger
	Now        time.Time
	errorCh    chan error
	finishCh   chan struct{}
	once       *sync.Once
	addresses  []string
	finishflag bool
}

func CreateCommitObserver(
	channel string,
	node basic.Node,
	crypto *basic.CryptoImpl,
	logger *log.Logger,
	n int,
	config basic.Config,
	errorCh chan error,
	finishCh chan struct{},
	once *sync.Once,
	finishflag bool) (*CommitObserver, error) {
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
	_, err = deliverer.Recv()
	if err != nil {
		return nil, err
	}
	addresses := make([]string, 0)
	for _, v := range config.Committers {
		addresses = append(addresses, v.Addr)
	}
	return &CommitObserver{d: deliverer,
		n:          n,
		logger:     logger,
		errorCh:    errorCh,
		finishCh:   finishCh,
		addresses:  addresses,
		once:       once,
		finishflag: finishflag,
	}, nil
}

func (o *CommitObserver) Start() {
	o.Now = time.Now()

	o.logger.Debugf("start observer for orderer")
	n := 0
	for {
		r, err := o.d.Recv()
		if err != nil {
			o.errorCh <- err
		}
		if r == nil {
			panic("Received nil message, but expect a valid block instead. You could look into your peer logs for more info")
		}
		block := r.GetBlock()
		tx := len(block.Data.Data)
		n += tx
		fmt.Printf("From Orderer Time %8.2fs\tBlock %6d\t Tx %6d\n", time.Since(o.Now).Seconds(), block.Header.Number, tx)
		for _, data := range block.Data.Data {
			txID := ""
			env, err := protoutil.GetEnvelopeFromBlock(data)
			if err != nil {
				continue
			}
			payload, err := protoutil.UnmarshalPayload(env.Payload)
			if err != nil {
				continue
			}
			chdr, err := protoutil.UnmarshalChannelHeader(payload.Header.ChannelHeader)
			if err != nil {
				continue
			}
			if common.HeaderType(chdr.Type) == common.HeaderType_ENDORSER_TRANSACTION {
				txID = chdr.TxId
			}
			if txID != "" {
				tapeSpan := basic.GetGlobalSpan()
				tapeSpan.FinishWithMap(txID, "", basic.CONSESUS)
				var span opentracing.Span
				if basic.GetMod() == infra.FULLPROCESS {
					Global_Span := tapeSpan.GetSpan(txID, "", basic.TRANSCATION)
					span = tapeSpan.SpanIntoMap(txID, "", basic.COMMIT_AT_ALL_PEERS, Global_Span)
				} else {
					span = tapeSpan.SpanIntoMap(txID, "", basic.COMMIT_AT_ALL_PEERS, nil)
				}
				tapeSpan.SpanIntoMap(txID, "", basic.COMMIT_AT_NETWORK, span)
				if basic.GetMod() != infra.COMMIT {
					for _, v := range o.addresses {
						tapeSpan.SpanIntoMap(txID, v, basic.COMMIT_AT_PEER, span)
					}
				}
				basic.LogEvent(o.logger, txID, "BlockFromOrderer")
			}
		}
		if o.n > 0 && o.finishflag {
			if n >= o.n {
				// consider with multiple threads need close this channel, need a once here to avoid channel been closed in multiple times
				o.once.Do(func() {
					close(o.finishCh)
				})
				return
			}
		}
	}
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
