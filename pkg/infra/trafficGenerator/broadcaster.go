package trafficGenerator

import (
	"context"
	"io"

	"github.com/hyperledger-twgc/tape/pkg/infra"
	"github.com/hyperledger-twgc/tape/pkg/infra/basic"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Broadcasters struct {
	workers []*Broadcaster
	ctx     context.Context
	envs    chan *basic.TracingEnvelope
	errorCh chan error
}

type Broadcaster struct {
	c      orderer.AtomicBroadcast_BroadcastClient
	logger *log.Logger
}

func CreateBroadcasters(ctx context.Context, envs chan *basic.TracingEnvelope, errorCh chan error, config basic.Config, logger *log.Logger) (*Broadcasters, error) {
	var workers []*Broadcaster
	for i := 0; i < config.NumOfConn; i++ {
		broadcaster, err := CreateBroadcaster(ctx, config.Orderer, logger)
		if err != nil {
			return nil, err
		}
		workers = append(workers, broadcaster)
	}

	return &Broadcasters{
		workers: workers,
		ctx:     ctx,
		envs:    envs,
		errorCh: errorCh,
	}, nil
}

func (bs Broadcasters) Start() {
	for _, b := range bs.workers {
		go b.StartDraining(bs.errorCh)
		go b.Start(bs.ctx, bs.envs, bs.errorCh)
	}
}

func CreateBroadcaster(ctx context.Context, node basic.Node, logger *log.Logger) (*Broadcaster, error) {
	client, err := basic.CreateBroadcastClient(ctx, node, logger)
	if err != nil {
		return nil, err
	}

	return &Broadcaster{c: client, logger: logger}, nil
}

func (b *Broadcaster) Start(ctx context.Context, envs <-chan *basic.TracingEnvelope, errorCh chan error) {
	b.logger.Debugf("Start sending broadcast")
	for {
		select {
		case e := <-envs:
			//b.logger.Debugf("Sending broadcast envelop")
			tapeSpan := basic.GetGlobalSpan()
			span := tapeSpan.MakeSpan(e.TxId, "", basic.BROADCAST, e.Span)
			err := b.c.Send(e.Env)
			if err != nil {
				errorCh <- err
			}
			span.Finish()
			e.Span.Finish()
			if basic.GetMod() == infra.FULLPROCESS {
				Global_Span := tapeSpan.GetSpan(e.TxId, "", basic.TRANSCATION)
				tapeSpan.SpanIntoMap(e.TxId, "", basic.CONSESUS, Global_Span)
			} else {
				tapeSpan.SpanIntoMap(e.TxId, "", basic.CONSESUS, nil)
			}

			e = nil
			// end of transcation
		case <-ctx.Done():
			return
		}
	}
}

func (b *Broadcaster) StartDraining(errorCh chan error) {
	for {
		res, err := b.c.Recv()
		if err != nil {
			if err == io.EOF {
				return
			}
			b.logger.Errorf("recv broadcast err: %+v, status: %+v\n", err, res)
			return
		}

		if res.Status != common.Status_SUCCESS {
			errorCh <- errors.Errorf("recv errouneous status %s", res.Status)
			return
		}
	}
}
