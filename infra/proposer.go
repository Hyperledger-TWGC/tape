package infra

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/peer"
)

type Proposers struct {
	workers []*Proposer

	client int
	index  uint64
}

func CreateProposers(conn, client int, addrs []string, crypto *Crypto) *Proposers {
	ps := make([]*Proposer, conn)
	for i := 0; i < conn; i++ {
		ps[i] = CreateProposer(addrs, crypto)
	}

	return &Proposers{workers: ps, client: client}
}

func (ps *Proposers) Start(signed, processed chan *Elecments, done <-chan struct{}) {
	fmt.Printf("Start sending transactions...\n\n")
	for _, p := range ps.workers {
		for i := 0; i < ps.client; i++ {
			go p.Start(signed, processed, done)
		}
	}
}

type Proposer struct {
	e []peer.EndorserClient
}

func CreateProposer(addrs []string, crypto *Crypto) *Proposer {
	endorser, err := CreateEndorserClient(addrs, crypto.TLSCACerts)
	if err != nil {
		panic(err)
	}

	return &Proposer{e: endorser}
}

func (p *Proposer) Start(signed, processed chan *Elecments, done <-chan struct{}) {
	for {
		select {
		case s := <-signed:
			endorsment := make([]*peer.ProposalResponse, len(p.e))
			skipper := false
			// add skipper to skip error, uncompleted proposal
			wg := sync.WaitGroup{}
			wg.Add(len(p.e))
			for n, _ := range p.e {
				go func(n int) {
					//ref golang pattern do loop in parallel
					//http://www.golangpatterns.info/concurrency/semaphores
					r, err := p.e[n].ProcessProposal(context.Background(), s.SignedProp)
					if err != nil || r.Response.Status < 200 || r.Response.Status >= 400 {
						fmt.Printf("Err processing proposal: %s, status: %d\n", err, r.Response.Status)
						skipper = true
					} else {
						endorsment[n] = r
					}
					wg.Done()
				}(n)
			}
			wg.Wait()
			if !skipper {
				s.Response = endorsment
				processed <- s
			}
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
