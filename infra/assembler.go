package infra

import (
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type Elecments struct {
	Proposal   *peer.Proposal
	SignedProp *peer.SignedProposal
	Response   *peer.ProposalResponse
	Envelope   *common.Envelope
}

type Assembler struct {
	Signer *Crypto
}

func (a *Assembler) assemble(e *Elecments) *Elecments {
	env, err := CreateSignedTx(e.Proposal, a.Signer, e.Response)
	if err != nil {
		panic(err)
	}
	e.Envelope = env
	return e
}

func (a *Assembler) sign(e *Elecments) *Elecments {
	sprop, err := SignProposal(e.Proposal, a.Signer)
	if err != nil {
		panic(err)
	}
	e.SignedProp = sprop
	return e
}

func (a *Assembler) StartSigner(raw, signed chan *Elecments, done <-chan struct{}) {
	for {
		select {
		case r := <-raw:
			signed <- a.sign(r)

		case <-done:
			return
		}
	}
}

func (a *Assembler) StartIntegrator(processed, envs chan *Elecments, done <-chan struct{}) {
	for {
		select {
		case p := <-processed:
			envs <- a.assemble(p)
		case <-done:
			return
		}
	}
}
