package trafficGenerator

import (
	"tape/internal/fabric/protoutil"
	"tape/pkg/infra"
	"tape/pkg/infra/basic"

	"github.com/hyperledger/fabric-protos-go/common"
)

type FackEnvelopGenerator struct {
	Num     int
	Burst   int
	R       float64
	Config  basic.Config
	Crypto  infra.Crypto
	Envs    chan *common.Envelope
	ErrorCh chan error
}

var nonce = []byte("nonce-abc-12345")
var data = []byte("data")

func (initiator *FackEnvelopGenerator) Start() {
	i := 0
	for {
		if initiator.Num > 0 {
			if i == initiator.Num {
				return
			}
			i++
		}
		creator, _ := initiator.Crypto.Serialize()
		txid := protoutil.ComputeTxID(nonce, creator)
		payloadBytes, _ := protoutil.GetBytesPayload(&common.Payload{
			Header: &common.Header{
				ChannelHeader: protoutil.MarshalOrPanic(&common.ChannelHeader{
					Type:      int32(common.HeaderType_ENDORSER_TRANSACTION),
					ChannelId: initiator.Config.Channel,
					TxId:      txid,
					Epoch:     uint64(0),
				}),
				SignatureHeader: protoutil.MarshalOrPanic(&common.SignatureHeader{
					Creator: creator,
					Nonce:   nonce,
				}),
			},
			Data: data,
		})

		signature, _ := initiator.Crypto.Sign(payloadBytes)

		initiator.Envs <- &common.Envelope{
			Payload:   payloadBytes,
			Signature: signature,
		}
	}
}
