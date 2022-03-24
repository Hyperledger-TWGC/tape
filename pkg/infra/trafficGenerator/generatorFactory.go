package trafficGenerator

import (
	"context"

	"github.com/hyperledger-twgc/tape/pkg/infra"
	"github.com/hyperledger-twgc/tape/pkg/infra/basic"

	log "github.com/sirupsen/logrus"
)

type TrafficGenerator struct {
	ctx          context.Context
	crypto       infra.Crypto
	raw          chan *basic.TracingProposal
	signed       []chan *basic.Elements
	envs         chan *basic.TracingEnvelope
	processed    chan *basic.Elements
	config       basic.Config
	num          int
	burst        int
	signerNumber int
	parallel     int
	rate         float64
	logger       *log.Logger
	errorCh      chan error
}

func NewTrafficGenerator(ctx context.Context, crypto infra.Crypto, envs chan *basic.TracingEnvelope, raw chan *basic.TracingProposal, processed chan *basic.Elements, signed []chan *basic.Elements, config basic.Config, num int, burst, signerNumber, parallel int, rate float64, logger *log.Logger, errorCh chan error) *TrafficGenerator {
	return &TrafficGenerator{
		ctx:          ctx,
		crypto:       crypto,
		raw:          raw,
		signed:       signed,
		envs:         envs,
		processed:    processed,
		config:       config,
		num:          num,
		parallel:     parallel,
		burst:        burst,
		signerNumber: signerNumber,
		rate:         rate,
		logger:       logger,
		errorCh:      errorCh,
	}
}

// table        | proposal boradcaster fake
// full process |  1          1        0      6
// commit       |  0          1        1      3
// query        |  1          0        0      4
func (t *TrafficGenerator) CreateGeneratorWorkers(mode int) ([]infra.Worker, error) {
	generator_workers := make([]infra.Worker, 0)
	// if create proposers int/4 = 1
	if mode/infra.PROPOSALFILTER == 1 {
		proposers, err := CreateProposers(t.ctx, t.signed, t.processed, t.config, t.logger)
		if err != nil {
			return generator_workers, err
		}
		generator_workers = append(generator_workers, proposers)
		assembler := &Assembler{Signer: t.crypto, Ctx: t.ctx, Raw: t.raw, Signed: t.signed, ErrorCh: t.errorCh, Logger: t.logger}
		Integrator := &Integrator{Signer: t.crypto, Ctx: t.ctx, Processed: t.processed, Envs: t.envs, ErrorCh: t.errorCh, Logger: t.logger}
		for i := 0; i < t.signerNumber; i++ {
			generator_workers = append(generator_workers, assembler)
			generator_workers = append(generator_workers, Integrator)
		}
	}
	// if boradcaster int mod 3 = 0
	if mode%infra.COMMITFILTER == 0 {
		broadcaster, err := CreateBroadcasters(t.ctx, t.envs, t.errorCh, t.config, t.logger)
		if err != nil {
			return generator_workers, err
		}
		generator_workers = append(generator_workers, broadcaster)
	}
	// if not fake int mod 2 = 0
	for i := 0; i < t.parallel; i++ {
		if mode%infra.QUERYFILTER == 0 {
			Initiator := &Initiator{Num: t.num, Burst: t.burst, R: t.rate, Config: t.config, Crypto: t.crypto, Logger: t.logger, Raw: t.raw, ErrorCh: t.errorCh}
			generator_workers = append(generator_workers, Initiator)
		} else {
			// if fake int mod 2 = 1
			fackEnvelopGenerator := &FackEnvelopGenerator{Num: t.num, Burst: t.burst, R: t.rate, Config: t.config, Crypto: t.crypto, Envs: t.envs, ErrorCh: t.errorCh}
			generator_workers = append(generator_workers, fackEnvelopGenerator)
		}
	}
	return generator_workers, nil
}
