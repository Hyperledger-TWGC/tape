package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func Process(configPath string, num int, burst int, rate float64, logger *log.Logger) error {
	config, err := LoadConfig(configPath)
	if err != nil {
		return err
	}
	crypto, err := config.LoadCrypto()
	if err != nil {
		return err
	}
	raw := make(chan *Elements, burst)
	signed := make([]chan *Elements, len(config.Endorsers))
	processed := make(chan *Elements, burst)
	envs := make(chan *Elements, burst)
	done := make(chan struct{})
	blockCh := make(chan *peer.FilteredBlock)
	finishCh := make(chan struct{})
	errorCh := make(chan error, burst)
	assembler := &Assembler{Signer: crypto}
	blockCollector, err := NewBlockCollector(config.CommitThreshold, len(config.Committers))
	if err != nil {
		return errors.Wrap(err, "failed to create block collector")
	}
	for i := 0; i < len(config.Endorsers); i++ {
		signed[i] = make(chan *Elements, burst)
	}

	for i := 0; i < 5; i++ {
		go assembler.StartSigner(raw, signed, errorCh, done)
		go assembler.StartIntegrator(processed, envs, errorCh, done)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	proposers, err := CreateProposers(config.NumOfConn, config.ClientPerConn, config.Endorsers, logger)
	if err != nil {
		return err
	}
	proposers.Start(ctx, signed, processed, done, config)

	broadcaster, err := CreateBroadcasters(ctx, config.NumOfConn, config.Orderer, logger)
	if err != nil {
		return err
	}
	broadcaster.Start(ctx, envs, errorCh, done)

	observers, err := CreateObservers(ctx, config.Channel, config.Committers, crypto, logger)
	if err != nil {
		return err
	}

	start := time.Now()

	go blockCollector.Start(ctx, blockCh, finishCh, num, time.Now(), true)
	go observers.Start(errorCh, blockCh, start)
	go StartCreateProposal(num, burst, rate, config, crypto, raw, errorCh, logger)

	for {
		select {
		case err = <-errorCh:
			return err
		case <-finishCh:
			duration := time.Since(start)
			close(done)

			logger.Infof("Completed processing transactions.")
			fmt.Printf("tx: %d, duration: %+v, tps: %f\n", num, duration, float64(num)/duration.Seconds())
			return nil
		}
	}
}
