package infra

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

const loglevel = "TAPE_LOGLEVEL"

var (
	app = kingpin.New("tape", "A performance test tool for Hyperledger Fabric")

	run = app.Command("run", "Start the tape program").Default()
	con = run.Flag("config", "Path to config file").Required().Short('c').String()
	num = run.Flag("number", "Number of tx for shot").Required().Short('n').Int()
)

func Main() {
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
	kingpin.MustParse(app.Parse(os.Args[1:]))

	config, err := LoadConfig(*con)
	if err != nil {
		return err
	}
	crypto, err := config.LoadCrypto()
	if err != nil {
		return err
	}
	raw := make(chan *Elements, 100)
	signed := make([]chan *Elements, len(config.Endorsers))
	processed := make(chan *Elements, 10)
	envs := make(chan *Elements, 10)
	done := make(chan struct{})
	finishCh := make(chan struct{})
	errorCh := make(chan error, 10)
	assember := &Assembler{Signer: crypto}

	for i := 0; i < len(config.Endorsers); i++ {
		signed[i] = make(chan *Elements, 10)
	}

	for i := 0; i < 5; i++ {
		go assember.StartSigner(raw, signed, errorCh, done)
		go assember.StartIntegrator(processed, envs, errorCh, done)
	}

	proposor, err := CreateProposers(config.NumOfConn, config.ClientPerConn, config.Endorsers, logger)
	if err != nil {
		return err
	}
	proposor.Start(signed, processed, done, config)

	broadcaster, err := CreateBroadcasters(config.NumOfConn, config.Orderer, logger)
	if err != nil {
		return err
	}
	broadcaster.Start(envs, errorCh, done)

	observer, err := CreateObserver(config.Channel, config.Committer, crypto, logger)
	if err != nil {
		return err
	}

	start := time.Now()
	go observer.Start(*num, errorCh, finishCh, start)
	go func() {
		for i := 0; i < *num; i++ {
			prop, err := CreateProposal(
				crypto,
				config.Channel,
				config.Chaincode,
				config.Version,
				config.Args...,
			)
			if err != nil {
				errorCh <- errors.Wrapf(err, "error creating proposal")
				return
			}
			raw <- &Elements{Proposal: prop}
		}
	}()

	for {
		select {
		case err = <-errorCh:
			return err
		case <-finishCh:
			duration := time.Since(start)
			close(done)

			logger.Infof("Completed processing transactions.")
			fmt.Printf("tx: %d, duration: %+v, tps: %f\n", *num, duration, float64(*num)/duration.Seconds())
			return nil
		}
	}
}
