package infra

import (
	"fmt"
	"math/big"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var one = new(big.Int).SetInt64(1)

func Process(configPath string, num int, burst int, rate float64, logger *log.Logger) error {
	CreatedNum := new(big.Int).SetInt64(0)
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
	finishCh := make(chan struct{})
	errorCh := make(chan error, burst)
	assember := &Assembler{Signer: crypto}
	blockCollector, err := NewBlockCollector(config.CommitThreshold, len(config.Committers))
	if err != nil {
		return errors.Wrap(err, "failed to create block collector")
	}
	for i := 0; i < len(config.Endorsers); i++ {
		signed[i] = make(chan *Elements, burst)
	}

	for i := 0; i < 5; i++ {
		go assember.StartSigner(raw, signed, errorCh, done)
		go assember.StartIntegrator(processed, envs, errorCh, done)
	}

	proposors, err := CreateProposers(config.NumOfConn, config.ClientPerConn, config.Endorsers, logger)
	if err != nil {
		return err
	}
	proposors.Start(signed, processed, done, config)

	broadcaster, err := CreateBroadcasters(config.NumOfConn, config.Orderer, logger)
	if err != nil {
		return err
	}
	broadcaster.Start(envs, errorCh, done)

	observers, err := CreateObservers(config.Channel, config.Committers, crypto, logger)
	if err != nil {
		return err
	}

	start := time.Now()
	go observers.Start(num, errorCh, finishCh, start, blockCollector, done)
	go StartCreateProposal(num, burst, rate, config, crypto, raw, errorCh, CreatedNum, done, logger)
	for {
		// adding signal handler
		select {
		case err = <-errorCh:
			return err
		case <-finishCh:
			duration := time.Since(start)
			close(done)

			logger.Infof("Completed processing transactions.")
			// to do if num == 0 here
			// confirmed num from blockCollector
			fmt.Printf("tx: %d, duration: %+v, tps: %f\n", blockCollector.GetConfirmNum(), duration, float64(blockCollector.GetConfirmNum())/duration.Seconds())
			return nil
		}
	}
}
