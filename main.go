package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/guoger/stupid/infra"
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
	if len(os.Args) != 3 {
		fmt.Printf("Usage: stupid config.yaml 500\n")
		os.Exit(1)
	}
	config := infra.LoadConfig(os.Args[1])
	N, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	crypto := config.LoadCrypto()

	raw := make(chan *infra.Elements, 100)
	signed := make([]chan *infra.Elements, len(config.Endorsers))
	processed := make(chan *infra.Elements, 10)
	envs := make(chan *infra.Elements, 10)
	done := make(chan struct{})

	assember := &infra.Assembler{Signer: crypto}

	for i := 0; i < len(config.Endorsers); i++ {
		signed[i] = make(chan *infra.Elements, 10)
	}

	for i := 0; i < 5; i++ {
		go assember.StartSigner(raw, signed, done)
		go assember.StartIntegrator(processed, envs, done)
	}

	proposor := infra.CreateProposers(config.NumOfConn, config.ClientPerConn, config.Endorsers, crypto, logger)
	proposor.Start(signed, processed, done, config)

	broadcaster := infra.CreateBroadcasters(config.NumOfConn, config.Orderer.Addr, crypto, logger)
	broadcaster.Start(envs, done)

	observer := infra.CreateObserver(config.Committer.Addr, config.Channel, crypto, logger)

	start := time.Now()
	go observer.Start(N, start)

	for i := 0; i < N; i++ {
		prop := infra.CreateProposal(
			crypto,
			config.Channel,
			config.Chaincode,
			config.Version,
			config.Args...,
		)
		raw <- &infra.Elements{Proposal: prop}
	}

	observer.Wait()
	duration := time.Since(start)
	close(done)
	logger.Infof("Completed processing transactions.")
	fmt.Printf("tx: %d, duration: %+v, tps: %f\n", N, duration, float64(N)/duration.Seconds())
	os.Exit(0)
}
