package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/guoger/stupid/infra"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: stupid config.json 500\n")
		os.Exit(1)
	}

	config := infra.LoadConfig(os.Args[1])
	N, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	crypto := config.LoadCrypto()

	raw := make(chan *infra.Elecments, 100)
	signed := make(chan *infra.Elecments, 10)
	processed := make(chan *infra.Elecments, 10)
	envs := make(chan *infra.Elecments, 10)
	done := make(chan struct{})

	assember := &infra.Assembler{Signer: crypto}
	for i := 0; i < 5; i++ {
		go assember.StartSigner(raw, signed, done)
		go assember.StartIntegrator(processed, envs, done)
	}

	proposor := infra.CreateProposers(config.NumOfConn, config.ClientPerConn, config.PeerAddr, crypto)
	proposor.Start(signed, processed, done)

	broadcaster := infra.CreateBroadcasters(len(config.Channels)*config.NumOfConn, config.OrdererAddr, crypto)
	broadcaster.Start(envs, done)
	var observer *infra.Observer
	if config.EventAddr == "" {
		observer = infra.CreateObserver(config.PeerAddr, config.Channels, crypto)
	} else {
		observer = infra.CreateObserver(config.EventAddr, config.Channels, crypto)
	}

	start := time.Now()
	go observer.Start(N, start, len(config.Channels))

	for i := 0; i < N; i++ {
		for _, channel := range config.Channels {
			prop := infra.CreateProposal(
				crypto,
				channel,
				config.Chaincode,
				config.Args...,
			)
			raw <- &infra.Elecments{Proposal: prop}
		}
	}

	observer.Wait()
	duration := time.Since(start)
	close(done)

	fmt.Printf("tx: %d, start time: %s, duration: %+v, tps: %f\n", len(config.Channels)*N, time.Unix(start.Unix(), 0).String(), duration, float64(len(config.Channels)*N)/duration.Seconds())
	os.Exit(0)
}
