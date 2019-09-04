package infra

import (
	"context"
	"time"

	"github.com/hyperledger/fabric/core/comm"
	"github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/peer"
)

func CreateGRPCClient(certs [][]byte) (*comm.GRPCClient, error) {
	config := comm.ClientConfig{}
	config.Timeout = 5 * time.Second
	config.SecOpts = &comm.SecureOptions{
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

func CreateEndorserClient(addr string, tlscacerts [][]byte) (peer.EndorserClient, error) {
	gRPCClient, err := CreateGRPCClient(tlscacerts)
	if err != nil {
		return nil, err
	}

	conn, err := gRPCClient.NewConnection(addr, "")
	if err != nil {
		return nil, err
	}

	return peer.NewEndorserClient(conn), nil
}

func CreateBroadcastClient(addr string, tlscacerts [][]byte) (orderer.AtomicBroadcast_BroadcastClient, error) {
	gRPCClient, err := CreateGRPCClient(tlscacerts)
	if err != nil {
		return nil, err
	}

	conn, err := gRPCClient.NewConnection(addr, "")
	if err != nil {
		return nil, err
	}

	return orderer.NewAtomicBroadcastClient(conn).Broadcast(context.Background())
}

func CreateDeliverFilteredClient(addr string, tlscacerts [][]byte) (peer.Deliver_DeliverFilteredClient, error) {
	gRPCClient, err := CreateGRPCClient(tlscacerts)
	if err != nil {
		return nil, err
	}

	conn, err := gRPCClient.NewConnection(addr, "")
	if err != nil {
		return nil, err
	}

	return peer.NewDeliverClient(conn).DeliverFiltered(context.Background())
}
