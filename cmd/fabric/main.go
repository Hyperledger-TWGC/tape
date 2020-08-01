package main

import (
	"fmt"
	"net"

	"github.com/guoger/stupid/mock"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:10086")
	if err != nil {
		panic(err)
	}

	fmt.Println("Start listening on localhost...")

	blockC := make(chan struct{}, 1000)

	p := &mock.Peer{
		BlkSize: 2000,
		TxC:     blockC,
	}

	o := &mock.Orderer{
		TxC: blockC,
	}

	grpcServer := grpc.NewServer()
	peer.RegisterEndorserServer(grpcServer, p)
	peer.RegisterDeliverServer(grpcServer, p)
	orderer.RegisterAtomicBroadcastServer(grpcServer, o)

	err = grpcServer.Serve(lis)
	if err != nil {
		panic(err)
	}
}
