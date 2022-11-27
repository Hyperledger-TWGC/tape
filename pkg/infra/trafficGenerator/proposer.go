package trafficGenerator

import (
	"context"
	"errors"

	"github.com/hyperledger-twgc/tape/pkg/infra/basic"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	log "github.com/sirupsen/logrus"
)

type Proposers struct {
	workers   [][]*Proposer
	logger    *log.Logger
	ctx       context.Context
	signed    []chan *basic.Elements
	processed chan *basic.Elements
	config    basic.Config
}

func CreateProposers(ctx context.Context, signed []chan *basic.Elements, processed chan *basic.Elements, config basic.Config, logger *log.Logger) (*Proposers, error) {
	var ps [][]*Proposer
	var err error
	//one proposer per connection per peer
	for _, node := range config.Endorsers {
		row := make([]*Proposer, config.NumOfConn)
		for j := 0; j < config.NumOfConn; j++ {
			row[j], err = CreateProposer(node, logger, config.Rule)
			if err != nil {
				return nil, err
			}
		}
		ps = append(ps, row)
	}

	return &Proposers{workers: ps, logger: logger, ctx: ctx, signed: signed, processed: processed, config: config}, nil
}

func (ps *Proposers) Start() {
	ps.logger.Infof("Start sending transactions.")
	for i := 0; i < len(ps.config.Endorsers); i++ {
		// peer connection should be config.ClientPerConn * config.NumOfConn
		for k := 0; k < ps.config.ClientPerConn; k++ {
			for j := 0; j < ps.config.NumOfConn; j++ {
				go ps.workers[i][j].Start(ps.ctx, ps.signed[i], ps.processed)
			}
		}
	}
}

type Proposer struct {
	e      peer.EndorserClient
	Addr   string
	logger *log.Logger
	Org    string
	rule   string
}

func CreateProposer(node basic.Node, logger *log.Logger, rule string) (*Proposer, error) {
	if len(rule) == 0 {
		return nil, errors.New("empty endorsement policy")
	}
	endorser, err := basic.CreateEndorserClient(node, logger)
	if err != nil {
		return nil, err
	}
	return &Proposer{e: endorser, Addr: node.Addr, logger: logger, Org: node.Org, rule: rule}, nil
}

func (p *Proposer) Start(ctx context.Context, signed, processed chan *basic.Elements) {
	tapeSpan := basic.GetGlobalSpan()
	for {
		select {
		case s := <-signed:
			//send sign proposal to peer for endorsement
			span := tapeSpan.MakeSpan(s.TxId, p.Addr, basic.ENDORSEMENT_AT_PEER, s.Span)
			r, err := p.e.ProcessProposal(ctx, s.SignedProp)
			if err != nil || r.Response.Status < 200 || r.Response.Status >= 400 {
				// end sending proposal
				if r == nil {
					p.logger.Errorf("Err processing proposal: %s, status: unknown, addr: %s \n", err, p.Addr)
				} else {
					p.logger.Errorf("Err processing proposal: %s, status: %d, message: %s, addr: %s \n", err, r.Response.Status, r.Response.Message, p.Addr)
				}
				continue
			}
			span.Finish()
			s.Lock.Lock()
			// if prometheus
			// report read readlatency with peer in label
			basic.GetLatencyMap().ReportReadLatency(s.TxId, p.Addr)
			s.Responses = append(s.Responses, r)
			s.Orgs = append(s.Orgs, p.Org)
			rs, err := CheckPolicy(s, p.rule)
			if err != nil {
				p.logger.Errorf("Fails to check rule of endorsement %s \n", err)
			}
			if rs {
				//if len(s.Responses) >= threshold { // from value upgrade to OPA logic
				s.EndorsementSpan.Finish()
				// OPA
				// if already in processed queue or after, ignore
				// if not, send into process queue
				processed <- s
				basic.LogEvent(p.logger, s.TxId, "CompletedCollectEndorsement")
			}
			s.Lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
