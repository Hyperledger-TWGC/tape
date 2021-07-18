package trafficGenerator

import (
	"context"
	"tape/pkg/infra"
	"tape/pkg/infra/basic"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
)

type Initiator struct {
	Num     int
	Burst   int
	R       float64
	Config  basic.Config
	Crypto  infra.Crypto
	Raw     chan *peer.Proposal
	ErrorCh chan error
}

func (initiator *Initiator) Start() {
	limit := rate.Inf
	ctx := context.Background()
	if initiator.R > 0 {
		limit = rate.Limit(initiator.R)
	}
	limiter := rate.NewLimiter(limit, initiator.Burst)
	for i := 0; i < initiator.Num; i++ {
		prop, err := CreateProposal(
			initiator.Crypto,
			initiator.Config.Channel,
			initiator.Config.Chaincode,
			initiator.Config.Version,
			initiator.Config.Args...,
		)
		if err != nil {
			initiator.ErrorCh <- errors.Wrapf(err, "error creating proposal")
			return
		}

		if err = limiter.Wait(ctx); err != nil {
			initiator.ErrorCh <- errors.Wrapf(err, "error creating proposal")
			return
		}
		initiator.Raw <- prop
	}
}
