package trafficGenerator

import (
	"context"

	"tape/pkg/infra/basic"

	"tape/pkg/infra"

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
	sprop, err := SignProposal(p.Proposal, a.Signer)
	if err != nil {
		return nil, err
	}
	basic.LogEvent(a.Logger, p.TxId, "SignProposal")

	return &basic.Elements{TxId: p.TxId, SignedProp: sprop}, nil
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
