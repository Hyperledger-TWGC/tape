package mock

import (
	"fmt"
	"io"
	"net"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Orderer struct {
	Listener   net.Listener
	GrpcServer *grpc.Server
	cnt        uint64
	TxCs       []chan struct{}
	SelfC      chan struct{}
}

func (o *Orderer) Deliver(srv orderer.AtomicBroadcast_DeliverServer) error {
	_, err := srv.Recv()
	if err != nil {
		panic("expect no recv error")
	}
	_ = srv.Send(&orderer.DeliverResponse{})
	for range o.SelfC {
		o.cnt++
		if o.cnt%10 == 0 {
			_ = srv.Send(&orderer.DeliverResponse{
				Type: &orderer.DeliverResponse_Block{Block: NewBlock(10, nil)},
			})
		}
	}
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

		for _, c := range o.TxCs {
			c <- struct{}{}
		}
		o.SelfC <- struct{}{}

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
		SelfC:      make(chan struct{}),
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
	_ = o.GrpcServer.Serve(o.Listener)
}

// NewBlock constructs a block with no data and no metadata.
func NewBlock(seqNum uint64, previousHash []byte) *common.Block {
	block := &common.Block{}
	block.Header = &common.BlockHeader{}
	block.Header.Number = seqNum
	block.Header.PreviousHash = previousHash
	block.Header.DataHash = []byte{}
	block.Data = &common.BlockData{}

	var metadataContents [][]byte
	for i := 0; i < len(common.BlockMetadataIndex_name); i++ {
		metadataContents = append(metadataContents, []byte{})
	}
	block.Metadata = &common.BlockMetadata{Metadata: metadataContents}

	return block
}
