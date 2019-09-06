package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/guoger/stupid/infra"
)

func main() {
	// This declares a string flag, -config, point to a config file.
	C := flag.String("config", "config.json", "Config JSON file that contains configurations about network and test variables.")
	// This declares an int flag, -total, stored the number of proposals.
	N := flag.Int("total", 40000, "Total number of proposals to send.")
	// Parse the command line flags.
	flag.Parse()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Failure:", r)
			fmt.Println()
			flag.Usage()
			os.Exit(1)
		}
	}()

	config := infra.LoadConfig(*C)
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

	broadcaster := infra.CreateBroadcasters(config.NumOfConn, config.OrdererAddr, crypto)
	broadcaster.Start(envs, done)

	observer := infra.CreateObserver(config.PeerAddr, config.Channel, crypto)

	start := time.Now()
	go observer.Start(*N, start)

	for i := 0; i < *N; i++ {
		prop := infra.CreateProposal(
			crypto,
			config.Channel,
			config.Chaincode,
			config.Args...,
		)
		raw <- &infra.Elecments{Proposal: prop}
	}

	observer.Wait()
	duration := time.Since(start)
	close(done)

	fmt.Printf("tx: %d, duration: %+v, tps: %f\n", *N, duration, float64(*N)/duration.Seconds())
	os.Exit(0)
}
