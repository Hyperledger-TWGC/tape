package trafficGenerator

import (
	"context"
	"tape/pkg/infra"
	"tape/pkg/infra/basic"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	log "github.com/sirupsen/logrus"
)

func CreateGeneratorWorkers(ctx context.Context, crypto infra.Crypto, raw chan *peer.Proposal, signed []chan *basic.Elements, envs chan *common.Envelope, processed chan *basic.Elements, config basic.Config, num int, burst int, rate float64, logger *log.Logger, errorCh chan error) ([]infra.Worker, error) {
	generator_workers := make([]infra.Worker, 0)
	proposers, err := CreateProposers(ctx, signed, processed, config, logger)
	if err != nil {
		return generator_workers, err
	}
	generator_workers = append(generator_workers, proposers)

	assembler := &Assembler{Signer: crypto, Ctx: ctx, Raw: raw, Signed: signed, ErrorCh: errorCh}
	Integrator := &Integrator{Signer: crypto, Ctx: ctx, Processed: processed, Envs: envs, ErrorCh: errorCh}
	for i := 0; i < 5; i++ {
		generator_workers = append(generator_workers, assembler)
		generator_workers = append(generator_workers, Integrator)
	}

	broadcaster, err := CreateBroadcasters(ctx, envs, errorCh, config, logger)
	if err != nil {
		return generator_workers, err
	}
	generator_workers = append(generator_workers, broadcaster)

	Initiator := &Initiator{Num: num, Burst: burst, R: rate, Config: config, Crypto: crypto, Raw: raw, ErrorCh: errorCh}
	generator_workers = append(generator_workers, Initiator)
	return generator_workers, nil
}
