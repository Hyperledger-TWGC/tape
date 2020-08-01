package mock_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/guoger/stupid/mock"
	"github.com/guoger/stupid/pkg/infra"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// generate cert and key and populate them to files
func genCertKey(key, cert *os.File) error {
	priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"STUPID"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	err = pem.Encode(cert, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return err
	}

	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil
	}
	err = pem.Encode(key, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	if err != nil {
		return err
	}

	return nil
}

func TestRegresssion(t *testing.T) {
	certFile, err := ioutil.TempFile("", "cert-")
	require.NoError(t, err)
	keyFile, err := ioutil.TempFile("", "key-")
	require.NoError(t, err)
	defer func() {
		os.Remove(certFile.Name())
		os.Remove(keyFile.Name())
	}()

	err = genCertKey(keyFile, certFile)
	require.NoError(t, err)

	addr := "127.0.0.1:10086"

	lis, err := net.Listen("tcp", addr)
	require.NoError(t, err)

	blockC := make(chan struct{}, 1000)

	p := &mock.Peer{
		BlkSize: 10,
		TxC:     blockC,
	}

	o := &mock.Orderer{
		TxC: blockC,
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	peer.RegisterEndorserServer(grpcServer, p)
	peer.RegisterDeliverServer(grpcServer, p)
	orderer.RegisterAtomicBroadcastServer(grpcServer, o)

	go grpcServer.Serve(lis)

	node := infra.Node{Addr: addr}
	config := infra.Config{
		Endorsers:     []infra.Node{node},
		Committer:     node,
		Orderer:       node,
		Channel:       "test-channel",
		Chaincode:     "test-chaincode",
		Version:       "0.1",
		Args:          nil,
		MSPID:         "test-msp",
		PrivateKey:    keyFile.Name(),
		SignCert:      certFile.Name(),
		NumOfConn:     10,
		ClientPerConn: 10,
	}

	crypto := config.LoadCrypto()
	logger := logrus.New()
	N := 100

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

	proposor := infra.CreateProposers(config.NumOfConn, config.ClientPerConn, config.Endorsers, logger)
	proposor.Start(signed, processed, done, config)

	broadcaster := infra.CreateBroadcasters(config.NumOfConn, config.Orderer, logger)
	broadcaster.Start(envs, done)

	observer := infra.CreateObserver(config.Channel, config.Committer, crypto, logger)

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

	require.NotZero(t, float64(N)/duration.Seconds())
}
