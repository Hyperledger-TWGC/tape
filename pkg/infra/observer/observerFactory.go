package observer

import (
	"context"
	"tape/pkg/infra"
	"tape/pkg/infra/basic"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func CreateObserverWorkers(config basic.Config, crypto infra.Crypto, blockCh chan *AddressedBlock, logger *log.Logger, ctx context.Context, finishCh chan struct{}, num int, errorCh chan error) ([]infra.Worker, *Observers, error) {
	observer_workers := make([]infra.Worker, 0)
	blockCollector, err := NewBlockCollector(config.CommitThreshold, len(config.Committers), ctx, blockCh, finishCh, num, true)
	if err != nil {
		return observer_workers, nil, errors.Wrap(err, "failed to create block collector")
	}
	observer_workers = append(observer_workers, blockCollector)
	observers, err := CreateObservers(ctx, crypto, errorCh, blockCh, config, logger)
	if err != nil {
		return observer_workers, observers, err
	}
	observer_workers = append(observer_workers, observers)
	return observer_workers, observers, nil
}

func CreateEndorsementObserverWorkers(envs chan *common.Envelope, ctx context.Context, finishCh chan struct{}, num int, errorCh chan error, logger *log.Logger) ([]infra.Worker, *EndorseObserver, error) {
	observer_workers := make([]infra.Worker, 0)
	EndorseObserverWorker := CreateEndorseObserver(envs, num, finishCh, logger)
	observer_workers = append(observer_workers, EndorseObserverWorker)
	return observer_workers, EndorseObserverWorker, nil
}

func CreateCommitObserverWorkers(channel string,
	node basic.Node,
	cryptoImpl *basic.CryptoImpl,
	logger *log.Logger,
	n int,
	errorCh chan error,
	finishCh chan struct{},
) ([]infra.Worker, *CommitObserver, error) {
	observer_workers := make([]infra.Worker, 0)
	EndorseObserverWorker := CreateCommitObserver(channel,
		node,
		cryptoImpl,
		logger,
		n,
		errorCh,
		finishCh)
	observer_workers = append(observer_workers, EndorseObserverWorker)
	return observer_workers, EndorseObserverWorker, nil
}
