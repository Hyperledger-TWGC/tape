package trafficGenerator

import (
	"context"
	"tape/pkg/infra"
	"tape/pkg/infra/basic"
)

type Integrator struct {
	Signer    infra.Crypto
	Ctx       context.Context
	Processed chan *basic.Elements
	Envs      chan *basic.Elements
	ErrorCh   chan error
}

func (integrator *Integrator) assemble(e *basic.Elements) (*basic.Elements, error) {
	env, err := CreateSignedTx(e.Proposal, integrator.Signer, e.Responses)
	if err != nil {
		return nil, err
	}
	e.Envelope = env
	return e, nil
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
		case <-integrator.Ctx.Done():
			return
		}
	}
}
