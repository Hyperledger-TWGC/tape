package cmdImpl

import (
	"context"
	"tape/pkg/infra"
	"tape/pkg/infra/basic"
	"tape/pkg/infra/observer"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type cmdConfig struct {
	Config    basic.Config
	Crypto    infra.Crypto
	Raw       chan *peer.Proposal
	Signed    []chan *basic.Elements
	Processed chan *basic.Elements
	Envs      chan *common.Envelope
	BlockCh   chan *observer.AddressedBlock
	FinishCh  chan struct{}
	ErrorCh   chan error
	Ctx       context.Context
	cancel    context.CancelFunc
}

func CreateCmd(configPath string, num int, burst, signerNumber int, rate float64) (*cmdConfig, error) {
	config, err := basic.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	crypto, err := config.LoadCrypto()
	if err != nil {
		return nil, err
	}
	raw := make(chan *peer.Proposal, burst)
	signed := make([]chan *basic.Elements, len(config.Endorsers))
	processed := make(chan *basic.Elements, burst)
	envs := make(chan *common.Envelope, burst)

	blockCh := make(chan *observer.AddressedBlock)

	finishCh := make(chan struct{})
	errorCh := make(chan error, burst)
	ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	for i := 0; i < len(config.Endorsers); i++ {
		signed[i] = make(chan *basic.Elements, burst)
	}
	cmd := &cmdConfig{
		config,
		crypto,
		raw,
		signed,
		processed,
		envs,
		blockCh,
		finishCh,
		errorCh,
		ctx,
		cancel,
	}
	return cmd, nil
}
