package infra

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/core/comm"
)

func CreateGRPCClient(cert []byte) (*comm.GRPCClient, error) {
	var certs [][]byte
	if cert != nil {
		certs = append(certs, cert)
	}
	config := comm.ClientConfig{}
	config.Timeout = 5 * time.Second
	config.SecOpts = comm.SecureOptions{
		UseTLS:            false,
		RequireClientCert: false,
		ServerRootCAs:     certs,
	}

	if len(certs) > 0 {
		config.SecOpts.UseTLS = true
	}

	grpcClient, err := comm.NewGRPCClient(config)
	if err != nil {
		return nil, err
	}

	return grpcClient, nil
}

func CreateEndorserClient(addr string, tlscacert []byte) (peer.EndorserClient, error) {
	gRPCClient, err := CreateGRPCClient(tlscacert)
	if err != nil {
		return nil, err
	}

	conn, err := gRPCClient.NewConnection(addr, func(tlsConfig *tls.Config) { tlsConfig.InsecureSkipVerify = true })
	if err != nil {
		return nil, err
	}

	return peer.NewEndorserClient(conn), nil
}

func CreateBroadcastClient(addr string, tlscacert []byte) (orderer.AtomicBroadcast_BroadcastClient, error) {
	gRPCClient, err := CreateGRPCClient(tlscacert)
	if err != nil {
		return nil, err
	}

	conn, err := gRPCClient.NewConnection(addr, func(tlsConfig *tls.Config) { tlsConfig.InsecureSkipVerify = true })
	if err != nil {
		return nil, err
	}

	return orderer.NewAtomicBroadcastClient(conn).Broadcast(context.Background())
}

func CreateDeliverFilteredClient(addr string, tlscacert []byte) (peer.Deliver_DeliverFilteredClient, error) {
	gRPCClient, err := CreateGRPCClient(tlscacert)
	if err != nil {
		return nil, err
	}

	conn, err := gRPCClient.NewConnection(addr, func(tlsConfig *tls.Config) { tlsConfig.InsecureSkipVerify = true })
	if err != nil {
		return nil, err
	}

	return peer.NewDeliverClient(conn).DeliverFiltered(context.Background())
}
