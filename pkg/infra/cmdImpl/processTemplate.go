package cmdImpl

import (
	"context"
	"tape/pkg/infra"
	"tape/pkg/infra/basic"
	"tape/pkg/infra/observer"
	"tape/pkg/infra/trafficGenerator"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	log "github.com/sirupsen/logrus"
)

type CmdConfig struct {
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
	Generator *trafficGenerator.TrafficGenerator
}

func CreateCmd(configPath string, num int, burst, signerNumber int, rate float64, logger *log.Logger) (*CmdConfig, error) {
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
	mytrafficGenerator := trafficGenerator.NewTrafficGenerator(ctx,
		crypto,
		envs,
		raw,
		processed,
		signed,
		config,
		num,
		burst,
		signerNumber,
		rate,
		logger,
		errorCh)

	cmd := &CmdConfig{
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
		mytrafficGenerator,
	}
	return cmd, nil
}
