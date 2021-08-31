package observer

import (
	"context"
	"tape/pkg/infra"
	"tape/pkg/infra/basic"

	"github.com/hyperledger/fabric-protos-go/common"
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
	envs     chan *common.Envelope
	errorCh  chan error
}

func NewObserverFactory(config basic.Config, crypto infra.Crypto, blockCh chan *AddressedBlock, logger *log.Logger, ctx context.Context, finishCh chan struct{}, num int, envs chan *common.Envelope, errorCh chan error) *ObserverFactory {
	return &ObserverFactory{config,
		crypto,
		blockCh,
		logger,
		ctx,
		finishCh,
		num,
		envs,
		errorCh,
	}
}

func (of *ObserverFactory) CreateObserverWorkers(mode int) ([]infra.Worker, infra.ObserverWorker, error) {
	switch mode {
	case 4:
		return of.CreateEndorsementObserverWorkers()
	case 3:
		return of.CreateCommitObserverWorkers()
	default:
		return of.CreateFullProcessObserverWorkers()
	}
}

// 6
func (of *ObserverFactory) CreateFullProcessObserverWorkers() ([]infra.Worker, infra.ObserverWorker, error) {
	observer_workers := make([]infra.Worker, 0)
	blockCollector, err := NewBlockCollector(of.config.CommitThreshold, len(of.config.Committers), of.ctx, of.blockCh, of.finishCh, of.num, true)
	if err != nil {
		return observer_workers, nil, errors.Wrap(err, "failed to create block collector")
	}
	observer_workers = append(observer_workers, blockCollector)
	observers, err := CreateObservers(of.ctx, of.crypto, of.errorCh, of.blockCh, of.config, of.logger)
	if err != nil {
		return observer_workers, observers, err
	}
	observer_workers = append(observer_workers, observers)
	return observer_workers, observers, nil
}

// 4
func (of *ObserverFactory) CreateEndorsementObserverWorkers() ([]infra.Worker, infra.ObserverWorker, error) {
	observer_workers := make([]infra.Worker, 0)
	EndorseObserverWorker := CreateEndorseObserver(of.envs, of.num, of.finishCh, of.logger)
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
	EndorseObserverWorker, err := CreateCommitObserver(of.config.Channel,
		of.config.Orderer,
		cryptoImpl,
		of.logger,
		of.num,
		of.errorCh,
		of.finishCh)
	if err != nil {
		return nil, nil, err
	}
	observer_workers = append(observer_workers, EndorseObserverWorker)
	return observer_workers, EndorseObserverWorker, nil
}
