package observer

import (
	"context"
	"sync"

	"github.com/hyperledger-twgc/tape/pkg/infra"
	"github.com/hyperledger-twgc/tape/pkg/infra/basic"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type ObserverFactory struct {
	config   basic.Config
	crypto   infra.Crypto
	blockCh  chan *AddressedBlock
	logger   *log.Logger
	ctx      context.Context
	finishCh chan struct{}
	num      int
	parallel int
	envs     chan *basic.TracingEnvelope
	errorCh  chan error
}

func NewObserverFactory(config basic.Config,
	crypto infra.Crypto,
	blockCh chan *AddressedBlock,
	logger *log.Logger,
	ctx context.Context,
	finishCh chan struct{},
	num, parallel int,
	envs chan *basic.TracingEnvelope,
	errorCh chan error) *ObserverFactory {
	return &ObserverFactory{config,
		crypto,
		blockCh,
		logger,
		ctx,
		finishCh,
		num,
		parallel,
		envs,
		errorCh,
	}
}

func (of *ObserverFactory) CreateObserverWorkers(mode int) ([]infra.Worker, infra.ObserverWorker, error) {
	switch mode {
	case infra.ENDORSEMENT:
		return of.CreateEndorsementObserverWorkers()
	case infra.OBSERVER:
		return of.CreateFullProcessObserverWorkers()
	case infra.COMMIT:
		return of.CreateCommitObserverWorkers()
	default:
		return of.CreateFullProcessObserverWorkers()
	}
}

// 6
func (of *ObserverFactory) CreateFullProcessObserverWorkers() ([]infra.Worker, infra.ObserverWorker, error) {
	observer_workers := make([]infra.Worker, 0)
	total := of.parallel * of.num
	var once sync.Once
	blockCollector, err := NewBlockCollector(of.config.CommitThreshold, len(of.config.Committers), of.ctx, of.blockCh, of.finishCh, total, true, of.logger, &once, true)
	if err != nil {
		return observer_workers, nil, errors.Wrap(err, "failed to create block collector")
	}
	observer_workers = append(observer_workers, blockCollector)
	observers, err := CreateObservers(of.ctx, of.crypto, of.errorCh, of.blockCh, of.config, of.logger)
	if err != nil {
		return observer_workers, observers, err
	}
	observer_workers = append(observer_workers, observers)
	cryptoImpl, err := of.config.LoadCrypto()
	if err != nil {
		return observer_workers, observers, err
	}
	EndorseObserverWorker, err := CreateCommitObserver(of.config.Channel,
		of.config.Orderer,
		cryptoImpl,
		of.logger,
		total,
		of.config,
		of.errorCh,
		of.finishCh,
		&once,
		false)
	if err != nil {
		return nil, nil, err
	}
	observer_workers = append(observer_workers, EndorseObserverWorker)
	return observer_workers, observers, nil
}

// 4
func (of *ObserverFactory) CreateEndorsementObserverWorkers() ([]infra.Worker, infra.ObserverWorker, error) {
	observer_workers := make([]infra.Worker, 0)
	total := of.parallel * of.num
	var once sync.Once
	EndorseObserverWorker := CreateEndorseObserver(of.envs, total, of.finishCh, &once, of.logger)
	observer_workers = append(observer_workers, EndorseObserverWorker)
	return observer_workers, EndorseObserverWorker, nil
}

// 3
func (of *ObserverFactory) CreateCommitObserverWorkers() ([]infra.Worker, infra.ObserverWorker, error) {
	observer_workers := make([]infra.Worker, 0)
	cryptoImpl, err := of.config.LoadCrypto()
	if err != nil {
		return observer_workers, nil, err
	}
	var once sync.Once
	total := of.parallel * of.num
	EndorseObserverWorker, err := CreateCommitObserver(of.config.Channel,
		of.config.Orderer,
		cryptoImpl,
		of.logger,
		total,
		of.config,
		of.errorCh,
		of.finishCh,
		&once,
		true)
	if err != nil {
		return nil, nil, err
	}
	observer_workers = append(observer_workers, EndorseObserverWorker)
	return observer_workers, EndorseObserverWorker, nil
}
