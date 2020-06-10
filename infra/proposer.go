package infra

import (
	"context"
	"fmt"
	"io"

	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/peer"
)

type Proposers struct {
	workers [][]*Proposer
	//one proposer per connection per peer
	client int
}

func CreateProposers(conn, client int, addrs []string, crypto *Crypto) *Proposers {
	var ps [][]*Proposer
	//one proposer per connection per peer
	for _, addr := range addrs {
		row := make([]*Proposer, conn)
		for j := 0; j < conn; j++ {
			row[j] = CreateProposer(addr, crypto)
		}
		ps = append(ps, row)
	}

	return &Proposers{workers: ps, client: client}
}

func (ps *Proposers) Start(signed []chan *Elecments, processed chan *Elecments, done <-chan struct{}, config Config) {
	fmt.Printf("Start sending transactions...\n\n")
	for i := 0; i < len(config.PeerAddrs); i++ {
		for j := 0; j < config.NumOfConn; j++ {
			go ps.workers[i][j].Start(signed[i], processed, done, len(config.PeerAddrs))
		}
	}
}

type Proposer struct {
	e    peer.EndorserClient
	addr string
}

func CreateProposer(addr string, crypto *Crypto) *Proposer {
	endorser, err := CreateEndorserClient(addr, crypto.TLSCACerts)
	if err != nil {
		panic(err)
	}

	return &Proposer{e: endorser, addr: addr}
}

func (p *Proposer) Start(signed, processed chan *Elecments, done <-chan struct{}, threshold int) {
	for {
		select {
		case s := <-signed:
			//send sign proposal to peer for endorsement
			r, err := p.e.ProcessProposal(context.Background(), s.SignedProp)
			if err != nil || r.Response.Status < 200 || r.Response.Status >= 400 {
				fmt.Printf("Err processing proposal: %s, status: %d, addr: %s \n", err, r.Response.Status, p.addr)
				fmt.Println(r)
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

func CreateBroadcasters(conn int, addr string, crypto *Crypto) Broadcasters {
	bs := make(Broadcasters, conn)
	for i := 0; i < conn; i++ {
		bs[i] = CreateBroadcaster(addr, crypto)
	}

	return bs
}

func (bs Broadcasters) Start(envs <-chan *Elecments, done <-chan struct{}) {
	for _, b := range bs {
		go b.StartDraining()
		go b.Start(envs, done)
	}
}

type Broadcaster struct {
	c orderer.AtomicBroadcast_BroadcastClient
}

func CreateBroadcaster(addr string, crypto *Crypto) *Broadcaster {
	client, err := CreateBroadcastClient(addr, crypto.TLSCACerts)
	if err != nil {
		panic(err)
	}

	return &Broadcaster{c: client}
}

func (b *Broadcaster) Start(envs <-chan *Elecments, done <-chan struct{}) {
	for {
		select {
		case e := <-envs:
			err := b.c.Send(e.Envelope)
			if err != nil {
				fmt.Printf("Failed to broadcast env: %s\n", err)
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

			fmt.Printf("Recv broadcast err: %s, status: %+v\n", err, res)
			panic("bcast recv err")
		}

		if res.Status != common.Status_SUCCESS {
			fmt.Printf("Recv errouneous status: %s\n", res.Status)
			panic("bcast recv err")
		}

	}
}
