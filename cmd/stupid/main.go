package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/guoger/stupid/pkg/infra"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/protoutil"
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
	}

	proposor := infra.CreateProposers(config.NumOfConn, config.ClientPerConn, config.Endorsers, logger)
	proposor.Start(signed, processed, done, config)

	broadcaster := infra.CreateBroadcasters(config.NumOfConn, config.Orderer, logger)
	broadcaster.Start(envs, done)

	observer := infra.CreateObserver(config.Channel, config.Committer, crypto, logger)
	endorseObserver := infra.CreateEndorseObserver(processed, logger)
	start := time.Now()

	if config.ProcessFlag == infra.ProcessAll {
		for i := 0; i < 5; i++ {
			go assember.StartIntegrator(processed, envs, done)
		}
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
	}
	if config.ProcessFlag == infra.EndorseOnly {
		go endorseObserver.Start(N, start)

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
		endorseObserver.Wait()
	}
	if config.ProcessFlag == infra.EnvelopeOnly {
		envelopeObserver := infra.CreateCommitObserver(config.Channel, config.Orderer, crypto, logger)
		go envelopeObserver.Start(N, start)
		for i := 0; i < N; i++ {
			nonce := []byte("nonce-abc-12345")
			creator, _ := crypto.Serialize()
			txid := protoutil.ComputeTxID(nonce, creator)

			txType := common.HeaderType_ENDORSER_TRANSACTION
			chdr := &common.ChannelHeader{
				Type:      int32(txType),
				ChannelId: config.Channel,
				TxId:      txid,
				Epoch:     uint64(0),
			}

			shdr := &common.SignatureHeader{
				Creator: creator,
				Nonce:   nonce,
			}

			payload := &common.Payload{
				Header: &common.Header{
					ChannelHeader:   protoutil.MarshalOrPanic(chdr),
					SignatureHeader: protoutil.MarshalOrPanic(shdr),
				},
				Data: []byte("data"),
			}
			payloadBytes, _ := protoutil.GetBytesPayload(payload)

			signature, _ := crypto.Sign(payloadBytes)

			envelope := &common.Envelope{
				Payload:   payloadBytes,
				Signature: signature,
			}

			envs <- &infra.Elements{Envelope: envelope}
		}
		envelopeObserver.Wait()
	}
	duration := time.Since(start)
	close(done)
	logger.Infof("Completed processing transactions.")
	fmt.Printf("tx: %d, duration: %+v, tps: %f\n", N, duration, float64(N)/duration.Seconds())
	os.Exit(0)
}
