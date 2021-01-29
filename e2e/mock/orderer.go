package mock

import (
	"fmt"
	"io"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
)

type Orderer struct {
	cnt uint64
	TxC chan struct{}
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

		o.TxC <- struct{}{}

		err = srv.Send(&orderer.BroadcastResponse{Status: common.Status_SUCCESS})
		if err != nil {
			return err
		}
	}
}
