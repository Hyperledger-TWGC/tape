package trafficGenerator

import (
	"context"
	"tape/pkg/infra"
	"tape/pkg/infra/basic"

	"github.com/hyperledger/fabric-protos-go/common"
)

type Integrator struct {
	Signer    infra.Crypto
	Ctx       context.Context
	Processed chan *basic.Elements
	Envs      chan *common.Envelope
	ErrorCh   chan error
}

func (integrator *Integrator) assemble(e *basic.Elements) (*common.Envelope, error) {
	env, err := CreateSignedTx(e.SignedProp, integrator.Signer, e.Responses)
	if err != nil {
		return nil, err
	}
	return env, nil
}

func (integrator *Integrator) Start() {
	for {
		select {
		case p := <-integrator.Processed:
			e, err := integrator.assemble(p)
			if err != nil {
				integrator.ErrorCh <- err
				return
			}
			integrator.Envs <- e
			p = nil
		case <-integrator.Ctx.Done():
			return
		}
	}
}
