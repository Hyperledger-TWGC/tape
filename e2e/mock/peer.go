package mock

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/peer"
)

type Peer struct {
	BlkSize, txCnt uint64
	TxC            chan struct{}
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
	srv.Send(&peer.DeliverResponse{})

	for range p.TxC {
		p.txCnt++
		if p.txCnt%p.BlkSize == 0 {
			srv.Send(&peer.DeliverResponse{Type: &peer.DeliverResponse_FilteredBlock{
				FilteredBlock: &peer.FilteredBlock{FilteredTransactions: make([]*peer.FilteredTransaction, p.BlkSize)}}})
		}
	}

	return nil
}

func (p *Peer) DeliverWithPrivateData(peer.Deliver_DeliverWithPrivateDataServer) error {
	panic("Not implemented")
}
