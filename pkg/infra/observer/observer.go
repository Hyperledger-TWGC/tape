package observer

import (
	"context"
	"time"

	"github.com/hyperledger-twgc/tape/pkg/infra"
	"github.com/hyperledger-twgc/tape/pkg/infra/basic"
	"github.com/hyperledger-twgc/tape/pkg/infra/trafficGenerator"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Observers struct {
	workers []*Observer
	errorCh chan error
	blockCh chan *AddressedBlock
	ctx     context.Context
}

type Observer struct {
	index   int
	Address string
	d       peer.Deliver_DeliverFilteredClient
	logger  *log.Logger
}

type key string

const start key = "start"

func CreateObservers(ctx context.Context, crypto infra.Crypto, errorCh chan error, blockCh chan *AddressedBlock, config basic.Config, logger *log.Logger) (*Observers, error) {
	var workers []*Observer
	for i, node := range config.Committers {
		worker, err := CreateObserver(ctx, config.Channel, node, crypto, logger)
		if err != nil {
			return nil, err
		}
		worker.index = i
		workers = append(workers, worker)
	}
	return &Observers{
		workers: workers,
		errorCh: errorCh,
		blockCh: blockCh,
		ctx:     ctx,
	}, nil
}

func (o *Observers) Start() {
	//o.StartTime = time.Now()

	o.ctx = context.WithValue(o.ctx, start, time.Now())
	for i := 0; i < len(o.workers); i++ {
		go o.workers[i].Start(o.errorCh, o.blockCh, o.ctx.Value(start).(time.Time))
	}
}

func (o *Observers) GetTime() time.Time {
	return o.ctx.Value(start).(time.Time)
}

func CreateObserver(ctx context.Context, channel string, node basic.Node, crypto infra.Crypto, logger *log.Logger) (*Observer, error) {
	seek, err := trafficGenerator.CreateSignedDeliverNewestEnv(channel, crypto)
	if err != nil {
		return nil, err
	}

	deliverer, err := basic.CreateDeliverFilteredClient(ctx, node, logger)
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

	return &Observer{Address: node.Addr, d: deliverer, logger: logger}, nil
}

func (o *Observer) Start(errorCh chan error, blockCh chan<- *AddressedBlock, now time.Time) {
	o.logger.Debugf("start observer for peer %s", o.Address)

	for {
		r, err := o.d.Recv()
		if err != nil {
			errorCh <- err
		}

		if r == nil {
			errorCh <- errors.Errorf("received nil message, but expect a valid block instead. You could look into your peer logs for more info")
			return
		}

		fb := r.Type.(*peer.DeliverResponse_FilteredBlock)
		for _, b := range fb.FilteredBlock.FilteredTransactions {
			basic.LogEvent(o.logger, b.Txid, "CommitAtPeer")
			tapeSpan := basic.GetGlobalSpan()
			tapeSpan.FinishWithMap(b.Txid, o.Address, basic.COMMIT_AT_PEER)
		}
		o.logger.Debugf("receivedTime %8.2fs\tBlock %6d\tTx %6d\t Address %s\n", time.Since(now).Seconds(), fb.FilteredBlock.Number, len(fb.FilteredBlock.FilteredTransactions), o.Address)

		blockCh <- &AddressedBlock{fb.FilteredBlock, o.index, time.Since(now)}
	}
}
