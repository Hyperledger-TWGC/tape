package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/guoger/stupid/pkg/infra"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const loglevel = "STUPID_LOGLEVEL"

func main() {
	logger := log.New()
	logger.SetLevel(log.WarnLevel)
	if customerLevel, customerSet := os.LookupEnv(loglevel); customerSet {
		if lvl, err := log.ParseLevel(customerLevel); err == nil {
			logger.SetLevel(lvl)
		}
	}
	err := process(logger)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func process(logger *log.Logger) error {
	if len(os.Args) != 3 {
		return errors.Errorf("error input parameters for stupid: stupid config.yaml 500")
	}
	N, err := strconv.Atoi(os.Args[2])
	if err != nil {
		return errors.Errorf("error input parameters for stupid: stupid config.yaml 500")
	}
	config, err := infra.LoadConfig(os.Args[1])
	if err != nil {
		return err
	}
	crypto, err := config.LoadCrypto()
	if err != nil {
		return err
	}
	raw := make(chan *infra.Elements, 100)
	signed := make([]chan *infra.Elements, len(config.Endorsers))
	processed := make(chan *infra.Elements, 10)
	envs := make(chan *infra.Elements, 10)
	done := make(chan struct{})
	finishCh := make(chan struct{})
	errorCh := make(chan error, 10)
	assember := &infra.Assembler{Signer: crypto}

	for i := 0; i < len(config.Endorsers); i++ {
		signed[i] = make(chan *infra.Elements, 10)
	}

	for i := 0; i < 5; i++ {
		go assember.StartSigner(raw, signed, errorCh, done)
		go assember.StartIntegrator(processed, envs, errorCh, done)
	}

	proposor, err := infra.CreateProposers(config.NumOfConn, config.ClientPerConn, config.Endorsers, logger)
	if err != nil {
		return err
	}
	proposor.Start(signed, processed, done, config)

	broadcaster, err := infra.CreateBroadcasters(config.NumOfConn, config.Orderer, logger)
	if err != nil {
		return err
	}
	broadcaster.Start(envs, errorCh, done)

	observer, err := infra.CreateObserver(config.Channel, config.Committer, crypto, logger)
	if err != nil {
		return err
	}

	start := time.Now()
	go observer.Start(N, errorCh, finishCh, start)

	for i := 0; i < N; i++ {
		prop, err := infra.CreateProposal(
			crypto,
			config.Channel,
			config.Chaincode,
			config.Version,
			config.Args...,
		)
		if err != nil {
			errCP := errors.Wrapf(err, "error creating proposal")
			return errCP
		}
		raw <- &infra.Elements{Proposal: prop}
	}
	for {
		select {
		case err = <-errorCh:
			return err
		case <-finishCh:
			duration := time.Since(start)
			close(done)

			logger.Infof("Completed processing transactions.")
			fmt.Printf("tx: %d, duration: %+v, tps: %f\n", N, duration, float64(N)/duration.Seconds())
			return nil
		}
	}
}
