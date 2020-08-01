package mock

import (
	"context"
	"fmt"
	"io"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
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
	return nil
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

func (n *Peer) DeliverWithPrivateData(peer.Deliver_DeliverWithPrivateDataServer) error {
	panic("Not implemented")
	return nil
}

type Orderer struct {
	cnt uint64
	TxC chan struct{}
}

func (o *Orderer) Deliver(orderer.AtomicBroadcast_DeliverServer) error {
	panic("Not implemented")
	return nil
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

		o.TxC <- struct{}{}

		err = srv.Send(&orderer.BroadcastResponse{Status: common.Status_SUCCESS})
		if err != nil {
			return err
		}
	}
}
