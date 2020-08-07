package mock

import (
	"net"

	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"google.golang.org/grpc"
)

type Server struct {
	GrpcServer *grpc.Server
	Listener   net.Listener
}

func (s *Server) Start() {
	blockC := make(chan struct{}, 1000)

	p := &Peer{
		BlkSize: 10,
		TxC:     blockC,
	}

	o := &Orderer{
		TxC: blockC,
	}

	peer.RegisterEndorserServer(s.GrpcServer, p)
	peer.RegisterDeliverServer(s.GrpcServer, p)
	orderer.RegisterAtomicBroadcastServer(s.GrpcServer, o)

	err := s.GrpcServer.Serve(s.Listener)
	if err != nil {
		panic(err)
	}
}

func (s *Server) Stop() {
	s.GrpcServer.Stop()
	s.Listener.Close()
}
