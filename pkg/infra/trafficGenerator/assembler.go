package trafficGenerator

import (
	"context"

	"tape/pkg/infra/basic"

	"tape/pkg/infra"
)

type Assembler struct {
	Signer  infra.Crypto
	Ctx     context.Context
	Raw     chan *basic.Elements
	Signed  []chan *basic.Elements
	ErrorCh chan error
}

func (a *Assembler) sign(e *basic.Elements) (*basic.Elements, error) {
	sprop, err := SignProposal(e.Proposal, a.Signer)
	if err != nil {
		return nil, err
	}
	e.SignedProp = sprop

	return e, nil
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
		case <-a.Ctx.Done():
			return
		}
	}
}
