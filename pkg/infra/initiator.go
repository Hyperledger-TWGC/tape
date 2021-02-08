package infra

import (
	"context"
	"math/big"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

func StartCreateProposal(num int, burst int, r float64, config Config, crypto *Crypto, raw chan *Elements, errorCh chan error, CreatedNum *big.Int, done <-chan struct{}, logger *log.Logger) {
	limit := rate.Inf
	ctx := context.Background()
	if r > 0 {
		limit = rate.Limit(r)
	}
	limiter := rate.NewLimiter(limit, burst)
	var Num *big.Int
	if num > 0 {
		Num = new(big.Int).SetInt64(int64(num))
	}
	for {
		select {
		case <-done:
			close(raw)
			return
		default:
			prop, err := CreateProposal(
				crypto,
				config.Channel,
				config.Chaincode,
				config.Version,
				config.Args...,
			)
			if err != nil {
				errorCh <- errors.Wrapf(err, "error creating proposal")
				return
			}

			if err = limiter.Wait(ctx); err != nil {
				errorCh <- errors.Wrapf(err, "error creating proposal")
				return
			}

			raw <- &Elements{Proposal: prop}
			CreatedNum.Add(CreatedNum, one)
			if Num != nil && CreatedNum.Cmp(Num) == 0 {
				return
			}
		}
	}
}
