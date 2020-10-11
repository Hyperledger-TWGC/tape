package infra

import (
	"sync"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type Elements struct {
	Proposal   *peer.Proposal
	SignedProp *peer.SignedProposal
	Responses  []*peer.ProposalResponse
	lock       sync.Mutex
	Envelope   *common.Envelope
}

type Assembler struct {
	Signer *Crypto
}

func (a *Assembler) assemble(e *Elements) (*Elements, error) {
	env, err := CreateSignedTx(e.Proposal, a.Signer, e.Responses)
	if err != nil {
		return nil, err
	}
	e.Envelope = env
	return e, nil
}

func (a *Assembler) sign(e *Elements) (*Elements, error) {
	sprop, err := SignProposal(e.Proposal, a.Signer)
	if err != nil {
		return nil, err
	}
	e.SignedProp = sprop

	return e, nil
}

func (a *Assembler) StartSigner(raw chan *Elements, signed []chan *Elements, errorCh chan error, done <-chan struct{}) {
	for {
		select {
		case r := <-raw:
			t, err := a.sign(r)
			if err != nil {
				errorCh <- err
				return
			}
			for _, v := range signed {
				v <- t
			}
		case <-done:
			return
		}
	}
}

func (a *Assembler) StartIntegrator(processed, envs chan *Elements, errorCh chan error, done <-chan struct{}) {
	for {
		select {
		case p := <-processed:
			e, err := a.assemble(p)
			if err != nil {
				errorCh <- err
				return
			}
			envs <- e
		case <-done:
			return
		}
	}
}
