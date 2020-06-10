package infra

import (
	"sync"

	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
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

func (a *Assembler) assemble(e *Elements) *Elements {
	env, err := CreateSignedTx(e.Proposal, a.Signer, e.Responses)
	if err != nil {
		panic(err)
	}

	e.Envelope = env
	return e
}

func (a *Assembler) sign(e *Elements) *Elements {
	sprop, err := SignProposal(e.Proposal, a.Signer)
	if err != nil {
		panic(err)
	}

	e.SignedProp = sprop

	return e
}

func (a *Assembler) StartSigner(raw chan *Elements, signed []chan *Elements, done <-chan struct{}) {
	for {
		select {
		case r := <-raw:
			t := a.sign(r)
			for _, v := range signed {
				v <- t
			}
		case <-done:
			return
		}
	}
}

func (a *Assembler) StartIntegrator(processed, envs chan *Elements, done <-chan struct{}) {
	for {
		select {
		case p := <-processed:
			envs <- a.assemble(p)
		case <-done:
			return
		}
	}
}
