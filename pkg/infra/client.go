package infra

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/core/comm"
	"github.com/pkg/errors"
)

func CreateGRPCClient(node Node) (*comm.GRPCClient, error) {
	var certs [][]byte
	if node.TLSCACertByte != nil {
		certs = append(certs, node.TLSCACertByte)
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
		if len(node.TLSCAKey) > 0 && len(node.TLSCARoot) > 0 {
			config.SecOpts.RequireClientCert = true
			config.SecOpts.Certificate = node.TLSCACertByte
			config.SecOpts.Key = node.TLSCAKeyByte
			if node.TLSCARootByte != nil {
				config.SecOpts.ClientRootCAs = append(config.SecOpts.ClientRootCAs, node.TLSCARootByte)
			}
		}
	}

	grpcClient, err := comm.NewGRPCClient(config)
	//to do: unit test for this error, current fails to make case for this
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to %s", node.Addr)
	}

	return grpcClient, nil
}

func CreateEndorserClient(node Node) (peer.EndorserClient, error) {
	gRPCClient, err := CreateGRPCClient(node)
	if err != nil {
		return nil, err
	}

	conn, err := gRPCClient.NewConnection(node.Addr, func(tlsConfig *tls.Config) { tlsConfig.InsecureSkipVerify = true })
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to %s", node.Addr)
	}

	return peer.NewEndorserClient(conn), nil
}

func CreateBroadcastClient(node Node) (orderer.AtomicBroadcast_BroadcastClient, error) {
	gRPCClient, err := CreateGRPCClient(node)
	if err != nil {
		return nil, err
	}

	conn, err := gRPCClient.NewConnection(node.Addr, func(tlsConfig *tls.Config) { tlsConfig.InsecureSkipVerify = true })
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to %s", node.Addr)
	}

	return orderer.NewAtomicBroadcastClient(conn).Broadcast(context.Background())
}

func CreateDeliverFilteredClient(node Node) (peer.Deliver_DeliverFilteredClient, error) {
	gRPCClient, err := CreateGRPCClient(node)
	if err != nil {
		return nil, err
	}

	conn, err := gRPCClient.NewConnection(node.Addr, func(tlsConfig *tls.Config) { tlsConfig.InsecureSkipVerify = true })
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to %s", node.Addr)
	}

	return peer.NewDeliverClient(conn).DeliverFiltered(context.Background())
}
