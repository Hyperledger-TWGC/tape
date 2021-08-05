package cmdImpl

import (
	"context"
	"fmt"
	"tape/pkg/infra/basic"
	"tape/pkg/infra/observer"
	"tape/pkg/infra/trafficGenerator"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	log "github.com/sirupsen/logrus"
)

func Process(configPath string, num int, burst, signerNumber int, rate float64, logger *log.Logger) error {
	/*** variables ***/
	config, err := basic.LoadConfig(configPath)
	if err != nil {
		return err
	}
	crypto, err := config.LoadCrypto()
	if err != nil {
		return err
	}
	raw := make(chan *peer.Proposal, burst)
	signed := make([]chan *basic.Elements, len(config.Endorsers))
	processed := make(chan *basic.Elements, burst)
	envs := make(chan *common.Envelope, burst)

	blockCh := make(chan *observer.AddressedBlock)

	finishCh := make(chan struct{})
	errorCh := make(chan error, burst)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for i := 0; i < len(config.Endorsers); i++ {
		signed[i] = make(chan *basic.Elements, burst)
	}
	/*** workers ***/
	Observer_workers, Observers, err := observer.CreateObserverWorkers(config, crypto, blockCh, logger, ctx, finishCh, num, errorCh)
	if err != nil {
		return err
	}
	generator_workers, err := trafficGenerator.CreateGeneratorWorkers(ctx, crypto, raw, signed, envs, processed, config, num, burst, signerNumber, rate, logger, errorCh)
	if err != nil {
		return err
	}
	/*** start workers ***/
	for _, worker := range Observer_workers {
		go worker.Start()
	}
	for _, worker := range generator_workers {
		go worker.Start()
	}
	/*** waiting for complete ***/
	for {
		select {
		case err = <-errorCh:
			return err
		case <-finishCh:
			duration := time.Since(Observers.GetTime())
			logger.Infof("Completed processing transactions.")
			fmt.Printf("tx: %d, duration: %+v, tps: %f\n", num, duration, float64(num)/duration.Seconds())
			return nil
		}
	}
}
