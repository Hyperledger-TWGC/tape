package trafficGenerator

import (
	"context"

	"github.com/hyperledger-twgc/tape/pkg/infra/basic"

	"github.com/hyperledger-twgc/tape/pkg/infra"

	log "github.com/sirupsen/logrus"
)

type Assembler struct {
	Signer  infra.Crypto
	Ctx     context.Context
	Raw     chan *basic.TracingProposal
	Signed  []chan *basic.Elements
	ErrorCh chan error
	Logger  *log.Logger
}

func (a *Assembler) sign(p *basic.TracingProposal) (*basic.Elements, error) {
	tapeSpan := basic.GetGlobalSpan()
	span := tapeSpan.MakeSpan(p.TxId, "", basic.SIGN_PROPOSAL, p.Span)
	defer span.Finish()

	sprop, err := SignProposal(p.Proposal, a.Signer)
	if err != nil {
		return nil, err
	}
	basic.LogEvent(a.Logger, p.TxId, "SignProposal")
	EndorsementSpan := tapeSpan.MakeSpan(p.TxId, "", basic.ENDORSEMENT, p.Span)
	orgs := make([]string, 0)
	basic.GetLatencyMap().StartTracing(p.TxId)
	return &basic.Elements{TxId: p.TxId, SignedProp: sprop, Span: p.Span, EndorsementSpan: EndorsementSpan, Orgs: orgs}, nil
}

func (a *Assembler) Start() {
	for {
		select {
		case r := <-a.Raw:
			t, err := a.sign(r)
			if err != nil {
				a.ErrorCh <- err
				return
			}
			for _, v := range a.Signed {
				v <- t
			}
			r = nil
		case <-a.Ctx.Done():
			return
		}
	}
}
