package trafficGenerator

import (
	"context"

	"tape/pkg/infra/basic"

	"tape/pkg/infra"

	"github.com/opentracing/opentracing-go"
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
	span := opentracing.GlobalTracer().StartSpan("Sign Proposal", opentracing.ChildOf(p.Span.Context()), opentracing.Tag{Key: "txid", Value: p.TxId})
	defer span.Finish()
	sprop, err := SignProposal(p.Proposal, a.Signer)
	if err != nil {
		return nil, err
	}
	basic.LogEvent(a.Logger, p.TxId, "SignProposal")
	EndorsementSpan := opentracing.GlobalTracer().StartSpan("Endorsements", opentracing.ChildOf(p.Span.Context()), opentracing.Tag{Key: "txid", Value: p.TxId})
	return &basic.Elements{TxId: p.TxId, SignedProp: sprop, Span: p.Span, EndorsementSpan: EndorsementSpan}, nil
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
