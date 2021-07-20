package trafficGenerator

import (
	"context"

	"tape/pkg/infra/basic"

	"tape/pkg/infra"

	"github.com/hyperledger/fabric-protos-go/peer"
)

type Assembler struct {
	Signer  infra.Crypto
	Ctx     context.Context
	Raw     chan *peer.Proposal
	Signed  []chan *basic.Elements
	ErrorCh chan error
}

func (a *Assembler) sign(p *peer.Proposal) (*basic.Elements, error) {
	sprop, err := SignProposal(p, a.Signer)
	if err != nil {
		return nil, err
	}

	return &basic.Elements{SignedProp: sprop}, nil
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
