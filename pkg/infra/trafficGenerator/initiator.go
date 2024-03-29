package trafficGenerator

import (
	"context"

	"github.com/hyperledger-twgc/tape/pkg/infra"
	"github.com/hyperledger-twgc/tape/pkg/infra/basic"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type Initiator struct {
	Num     int
	Burst   int
	R       float64
	Config  basic.Config
	Crypto  infra.Crypto
	Logger  *log.Logger
	Raw     chan *basic.TracingProposal
	ErrorCh chan error
}

func (initiator *Initiator) Start() {
	limit := rate.Inf
	ctx := context.Background()
	if initiator.R > 0 {
		limit = rate.Limit(initiator.R)
	}
	limiter := rate.NewLimiter(limit, initiator.Burst)
	i := 0
	for {
		if initiator.Num > 0 {
			if i == initiator.Num {
				return
			}
			i++
		}
		prop, err := CreateProposal(
			initiator.Crypto,
			initiator.Logger,
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
