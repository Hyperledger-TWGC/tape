package cmdImpl

import (
	"context"
	"tape/pkg/infra/basic"
	"tape/pkg/infra/observer"
	"tape/pkg/infra/trafficGenerator"

	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
)

type CmdConfig struct {
	FinishCh        chan struct{}
	ErrorCh         chan error
	cancel          context.CancelFunc
	Generator       *trafficGenerator.TrafficGenerator
	Observerfactory *observer.ObserverFactory
}

func CreateCmd(configPath string, num int, burst, signerNumber, parallel int, rate float64, logger *log.Logger) (*CmdConfig, error) {
	config, err := basic.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	crypto, err := config.LoadCrypto()
	if err != nil {
		return nil, err
	}
	raw := make(chan *basic.TracingProposal, burst)
	signed := make([]chan *basic.Elements, len(config.Endorsers))
	processed := make(chan *basic.Elements, burst)
	envs := make(chan *basic.TracingEnvelope, burst)

	blockCh := make(chan *observer.AddressedBlock)

	finishCh := make(chan struct{})
	errorCh := make(chan error, burst)
	ctx, cancel := context.WithCancel(context.Background())

	tr, closer := basic.Init("tape")
	defer closer.Close()
	opentracing.SetGlobalTracer(tr)
	//defer cancel()
	for i := 0; i < len(config.Endorsers); i++ {
		signed[i] = make(chan *basic.Elements, burst)
	}

	Spans := make(map[string]opentracing.Span)
	Tspans := &basic.TracingSpans{
		Spans: Spans,
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
		parallel,
		rate,
		logger,
		errorCh)

	Observerfactory := observer.NewObserverFactory(
		config,
		crypto,
		blockCh,
		logger,
		ctx,
		finishCh,
		num,
		parallel,
		envs,
		Tspans,
		errorCh)
	cmd := &CmdConfig{
		finishCh,
		errorCh,
		cancel,
		mytrafficGenerator,
		Observerfactory,
	}
	return cmd, nil
}
