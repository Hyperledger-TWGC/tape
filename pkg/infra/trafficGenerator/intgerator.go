package trafficGenerator

import (
	"context"

	"github.com/hyperledger-twgc/tape/pkg/infra"
	"github.com/hyperledger-twgc/tape/pkg/infra/basic"

	log "github.com/sirupsen/logrus"
)

type Integrator struct {
	Signer    infra.Crypto
	Ctx       context.Context
	Processed chan *basic.Elements
	Envs      chan *basic.TracingEnvelope
	ErrorCh   chan error
	Logger    *log.Logger
}

func (integrator *Integrator) assemble(e *basic.Elements) (*basic.TracingEnvelope, error) {
	tapeSpan := basic.GetGlobalSpan()
	span := tapeSpan.MakeSpan(e.TxId, "", basic.SIGN_ENVELOP, e.Span)
	defer span.Finish()
	env, err := CreateSignedTx(e.SignedProp, integrator.Signer, e.Responses)
	// end integration proposal
	basic.LogEvent(integrator.Logger, e.TxId, "CreateSignedEnvelope")
	if err != nil {
		return nil, err
	}
	return &basic.TracingEnvelope{Env: env, TxId: e.TxId, Span: e.Span}, nil
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
