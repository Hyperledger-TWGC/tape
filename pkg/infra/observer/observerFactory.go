package observer

import (
	"context"
	"tape/pkg/infra"
	"tape/pkg/infra/basic"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func CreateObserverWorkers(config basic.Config, crypto infra.Crypto, blockCh chan *AddressedBlock, logger *log.Logger, ctx context.Context, finishCh chan struct{}, num int, errorCh chan error) ([]infra.Worker, error) {
	observer_workers := make([]infra.Worker, 0)
	blockCollector, err := NewBlockCollector(config.CommitThreshold, len(config.Committers), ctx, blockCh, finishCh, num, true)
	if err != nil {
		return observer_workers, errors.Wrap(err, "failed to create block collector")
	}
	observer_workers = append(observer_workers, blockCollector)
	observers, err := CreateObservers(ctx, crypto, errorCh, blockCh, config, logger)
	if err != nil {
		return observer_workers, err
	}
	observer_workers = append(observer_workers, observers)
	return observer_workers, nil
}
