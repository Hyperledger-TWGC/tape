package mock

import (
	"context"
	"net"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Peer struct {
	Listener       net.Listener
	GrpcServer     *grpc.Server
	BlkSize, txCnt uint64
	TxC            chan struct{}
	ctlCh          chan bool
}

func (p *Peer) ProcessProposal(context.Context, *peer.SignedProposal) (*peer.ProposalResponse, error) {
	return &peer.ProposalResponse{Response: &peer.Response{Status: 200}}, nil
}

func (p *Peer) Deliver(peer.Deliver_DeliverServer) error {
	panic("Not implemented")
}

func (p *Peer) DeliverFiltered(srv peer.Deliver_DeliverFilteredServer) error {
	_, err := srv.Recv()
	if err != nil {
		panic("expect no recv error")
	}
	_ = srv.Send(&peer.DeliverResponse{})
	txc := p.TxC
	for {
		select {
		case <-txc:
			p.txCnt++
			if p.txCnt%p.BlkSize == 0 {
				_ = srv.Send(&peer.DeliverResponse{Type: &peer.DeliverResponse_FilteredBlock{
					FilteredBlock: &peer.FilteredBlock{
						Number:               p.txCnt / p.BlkSize,
						FilteredTransactions: make([]*peer.FilteredTransaction, p.BlkSize)}}})
			}
		case pause := <-p.ctlCh:
			if pause {
				txc = nil
			} else {
				txc = p.TxC
			}
		}
	}
}

func (p *Peer) DeliverWithPrivateData(peer.Deliver_DeliverWithPrivateDataServer) error {
	panic("Not implemented")
}

func (p *Peer) Stop() {
	p.GrpcServer.Stop()
	p.Listener.Close()
}

func (p *Peer) Start() {
	_ = p.GrpcServer.Serve(p.Listener)
}

func (p *Peer) Addrs() string {
	return p.Listener.Addr().String()
}

func NewPeer(TxC chan struct{}, credentials credentials.TransportCredentials) (*Peer, error) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	ctlCh := make(chan bool)
	instance := &Peer{
		Listener:   lis,
		GrpcServer: grpc.NewServer(grpc.Creds(credentials)),
		BlkSize:    10,
		TxC:        TxC,
		ctlCh:      ctlCh,
	}

	peer.RegisterEndorserServer(instance.GrpcServer, instance)
	peer.RegisterDeliverServer(instance.GrpcServer, instance)

	return instance, nil
}

func (p *Peer) Pause() {
	p.ctlCh <- true
}

func (p *Peer) Unpause() {
	p.ctlCh <- false
}
