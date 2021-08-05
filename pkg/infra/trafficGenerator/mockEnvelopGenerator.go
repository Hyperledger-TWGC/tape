package trafficGenerator

import (
	"tape/internal/fabric/protoutil"
	"tape/pkg/infra"
	"tape/pkg/infra/basic"

	"github.com/hyperledger/fabric-protos-go/common"
)

type mockEnvelopGenerator struct {
	Num     int
	Burst   int
	R       float64
	Config  basic.Config
	Crypto  infra.Crypto
	Envs    chan *common.Envelope
	ErrorCh chan error
}

func (initiator *mockEnvelopGenerator) Start() {
	for i := 0; i < initiator.Num; i++ {
		nonce := []byte("nonce-abc-12345")
		creator, _ := initiator.Crypto.Serialize()
		txid := protoutil.ComputeTxID(nonce, creator)

		txType := common.HeaderType_ENDORSER_TRANSACTION
		chdr := &common.ChannelHeader{
			Type:      int32(txType),
			ChannelId: initiator.Config.Channel,
			TxId:      txid,
			Epoch:     uint64(0),
		}

		shdr := &common.SignatureHeader{
			Creator: creator,
			Nonce:   nonce,
		}

		payload := &common.Payload{
			Header: &common.Header{
				ChannelHeader:   protoutil.MarshalOrPanic(chdr),
				SignatureHeader: protoutil.MarshalOrPanic(shdr),
			},
			Data: []byte("data"),
		}
		payloadBytes, _ := protoutil.GetBytesPayload(payload)

		signature, _ := initiator.Crypto.Sign(payloadBytes)

		envelope := &common.Envelope{
			Payload:   payloadBytes,
			Signature: signature,
		}

		initiator.Envs <- envelope
	}
}
