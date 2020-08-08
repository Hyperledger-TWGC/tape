package mock

import (
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric/protoutil"
)

type Orderer struct {
	cnt uint64
	TxC chan struct{}
}

func (o *Orderer) Deliver(srv orderer.AtomicBroadcast_DeliverServer) error {

	_, err := srv.Recv()
	if err != nil {
		panic("expect no recv error")
	}
	for range o.TxC {
		o.cnt++
		if o.cnt%10 == 0 {
			srv.Send(&orderer.DeliverResponse{
				Type: &orderer.DeliverResponse_Block{Block: protoutil.NewBlock(10, nil)},
			})
		}
	}
	return nil
}

func (o *Orderer) Broadcast(srv orderer.AtomicBroadcast_BroadcastServer) error {
	for {
		o.TxC <- struct{}{}
		err := srv.Send(&orderer.BroadcastResponse{Status: common.Status_SUCCESS})
		if err != nil {
			return err
		}
	}
}
