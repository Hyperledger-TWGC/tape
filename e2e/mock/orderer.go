package mock

import (
	"fmt"
	"io"
	"net"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Orderer struct {
	Listener   net.Listener
	GrpcServer *grpc.Server
	cnt        uint64
	TxCs       []chan struct{}
}

func (o *Orderer) Deliver(orderer.AtomicBroadcast_DeliverServer) error {
	panic("Not implemented")
}

func (o *Orderer) Broadcast(srv orderer.AtomicBroadcast_BroadcastServer) error {
	for {
		_, err := srv.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			fmt.Println(err)
			return err
		}

		for _, c := range o.TxCs {
			c <- struct{}{}
		}

		err = srv.Send(&orderer.BroadcastResponse{Status: common.Status_SUCCESS})
		if err != nil {
			return err
		}
	}
}

func NewOrderer(txCs []chan struct{}, credentials credentials.TransportCredentials) (*Orderer, error) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	instance := &Orderer{
		Listener:   lis,
		GrpcServer: grpc.NewServer(grpc.Creds(credentials)),
		TxCs:       txCs,
	}
	orderer.RegisterAtomicBroadcastServer(instance.GrpcServer, instance)
	return instance, nil
}

func (o *Orderer) Stop() {
	o.GrpcServer.Stop()
	o.Listener.Close()
}

func (o *Orderer) Addrs() string {
	return o.Listener.Addr().String()
}

func (o *Orderer) Start() {
	o.GrpcServer.Serve(o.Listener)
}
