package infra

import (
	"context"
	"io"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	log "github.com/sirupsen/logrus"
)

type Proposers struct {
	workers [][]*Proposer
	//one proposer per connection per peer
	client int
	logger *log.Logger
}

func CreateProposers(conn, client int, nodes []Node, logger *log.Logger) *Proposers {
	var ps [][]*Proposer
	//one proposer per connection per peer
	for _, node := range nodes {
		row := make([]*Proposer, conn)
		for j := 0; j < conn; j++ {
			row[j] = CreateProposer(node, logger)
		}
		ps = append(ps, row)
	}

	return &Proposers{workers: ps, client: client, logger: logger}
}

func (ps *Proposers) Start(signed []chan *Elements, processed chan *Elements, done <-chan struct{}, config Config) {
	ps.logger.Infof("Start sending transactions.")
	for i := 0; i < len(config.Endorsers); i++ {
		// peer connection should be config.ClientPerConn * config.NumOfConn
		for k := 0; k < config.ClientPerConn; k++ {
			for j := 0; j < config.NumOfConn; j++ {
				go ps.workers[i][j].Start(signed[i], processed, done, len(config.Endorsers))
			}
		}
	}
}

type Proposer struct {
	e      peer.EndorserClient
	Addr   string
	logger *log.Logger
}

func CreateProposer(node Node, logger *log.Logger) *Proposer {
	endorser, err := CreateEndorserClient(node)
	if err != nil {
		panic(err)
	}

	return &Proposer{e: endorser, Addr: node.Addr, logger: logger}
}

func (p *Proposer) Start(signed, processed chan *Elements, done <-chan struct{}, threshold int) {
	for {
		select {
		case s := <-signed:
			//send sign proposal to peer for endorsement
			r, err := p.e.ProcessProposal(context.Background(), s.SignedProp)
			if err != nil || r.Response.Status < 200 || r.Response.Status >= 400 {
				if r == nil {
					p.logger.Errorf("Err processing proposal: %s, status: unknown, addr: %s \n", err, p.Addr)
				} else {
					p.logger.Errorf("Err processing proposal: %s, status: %d, addr: %s \n", err, r.Response.Status, p.Addr)
				}
				continue
			}
			s.lock.Lock()
			//collect for endorsement
			s.Responses = append(s.Responses, r)
			if len(s.Responses) >= threshold {
				processed <- s
			}
			s.lock.Unlock()
		case <-done:
			return
		}
	}
}

type Broadcasters []*Broadcaster

func CreateBroadcasters(conn int, orderer Node, logger *log.Logger) Broadcasters {
	bs := make(Broadcasters, conn)
	for i := 0; i < conn; i++ {
		bs[i] = CreateBroadcaster(orderer, logger)
	}

	return bs
}

func (bs Broadcasters) Start(envs <-chan *Elements, done <-chan struct{}) {
	for _, b := range bs {
		go b.StartDraining()
		go b.Start(envs, done)
	}
}

type Broadcaster struct {
	c      orderer.AtomicBroadcast_BroadcastClient
	logger *log.Logger
}

func CreateBroadcaster(node Node, logger *log.Logger) *Broadcaster {
	client, err := CreateBroadcastClient(node)
	if err != nil {
		panic(err)
	}

	return &Broadcaster{c: client, logger: logger}
}

func (b *Broadcaster) Start(envs <-chan *Elements, done <-chan struct{}) {
	b.logger.Debugf("Start sending broadcast")
	for {
		select {
		case e := <-envs:
			err := b.c.Send(e.Envelope)
			if err != nil {
				b.logger.Errorf("Failed to broadcast env: %s\n", err)
			}

		case <-done:
			return
		}
	}
}

func (b *Broadcaster) StartDraining() {
	for {
		res, err := b.c.Recv()
		if err != nil {
			if err == io.EOF {
				return
			}
			b.logger.Errorf("Recv broadcast err: %s, status: %+v\n", err, res)
			return
		}

		if res.Status != common.Status_SUCCESS {
			b.logger.Errorf("Recv errouneous status: %s\n", res.Status)
			panic("bcast recv err")
		}

	}
}
