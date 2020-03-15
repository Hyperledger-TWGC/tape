package infra

import (
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

const testAddress = "127.0.0.1:0"

var signed chan *Elements
var processed chan *Elements
var done chan struct{}

func beforeEach(size int) {
	signed = make(chan *Elements, size)
	processed = make(chan *Elements, size)
	done = make(chan struct{})
}

func TestSuccessEndorsement(t *testing.T) {
	beforeEach(10)
	srv := &mocks.MockEndorserServer{}
	defer srv.Stop()
	targetProposer := targetProposer(srv)
	go targetProposer.Start(signed, processed, done, 1)
	signProposal(signed, done, 1)
	afterEach()
	if len(processed) == 0 {
		t.Fatalf("it should be processed for all")
	}
}

func TestFailEndorsement(t *testing.T) {
	beforeEach(10)
	srv := &mocks.MockEndorserServer{ProposalError: fmt.Errorf("error")}
	defer srv.Stop()
	targetProposer := targetProposer(srv)
	go targetProposer.Start(signed, processed, done, 1)
	signProposal(signed, done, 1)
	afterEach()
	if len(processed) != 0 {
		t.Fatalf("it should not be processed for all")
	}
}

func Benchmark_Endorsement(b *testing.B) {
	b.StopTimer()
	beforeEach(b.N)
	srv := &mocks.MockEndorserServer{}
	targetProposer := targetProposer(srv)
	go targetProposer.Start(signed, processed, done, 1)
	defer srv.Stop()
	b.StartTimer()
	for i := 0; i < b.N; i++ { //use b.N for looping
		signProposal(signed, done, 1)
	}
}

func targetProposer(srv *mocks.MockEndorserServer) *Proposer {
	addr := srv.Start(testAddress)
	//defer srv.Stop()
	crypto := &Crypto{}
	return CreateProposer(addr, crypto)
}

func signProposal(signed chan *Elements, done chan struct{}, numbers int) {
	for i := 0; i < numbers; i++ {
		Prop := &Elements{SignedProp: &peer.SignedProposal{ProposalBytes: []byte("endorser"), Signature: []byte("dummy")}}
		signed <- Prop
	}
	time.Sleep(time.Duration(1) * time.Millisecond)
}

func afterEach() {
	close(done)
}
