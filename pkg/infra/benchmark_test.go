package infra

import (
	"net"
	"testing"

	"github.com/hyperledger/fabric-protos-go/peer"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"tape/e2e/mock"
)

func StartMockPeer() (*mock.Server, string) {
	lis, _ := net.Listen("tcp", "127.0.0.1:0")
	grpcServer := grpc.NewServer()
	mockPeer := &mock.Server{GrpcServer: grpcServer, Listener: lis}
	go mockPeer.Start()
	return mockPeer, lis.Addr().String()
}

func StartProposer(signed, processed chan *Elements, done chan struct{}, logger *log.Logger, threshold int, addr string) {
	peer := Node{
		Addr: addr,
	}
	Proposer, _ := CreateProposer(peer, logger)
	go Proposer.Start(signed, processed, done, threshold)
}

func benchmarkNPeer(concurrent int, b *testing.B) {
	processed := make(chan *Elements, 10)
	done := make(chan struct{})
	defer close(done)
	signeds := make([]chan *Elements, concurrent)
	for i := 0; i < concurrent; i++ {
		signeds[i] = make(chan *Elements, 10)
		mockpeer, mockpeeraddr := StartMockPeer()
		StartProposer(signeds[i], processed, done, nil, concurrent, mockpeeraddr)
		defer mockpeer.Stop()
	}

	b.ResetTimer()
	go func() {
		for i := 0; i < b.N; i++ {
			data := &Elements{SignedProp: &peer.SignedProposal{}}
			for _, s := range signeds {
				s <- data
			}
		}
	}()
	var n int
	for n < b.N {
		<-processed
		n++
	}
	b.StopTimer()
}

func BenchmarkPeerEndorsement1(b *testing.B) { benchmarkNPeer(1, b) }
func BenchmarkPeerEndorsement2(b *testing.B) { benchmarkNPeer(2, b) }
func BenchmarkPeerEndorsement4(b *testing.B) { benchmarkNPeer(4, b) }
func BenchmarkPeerEndorsement8(b *testing.B) { benchmarkNPeer(8, b) }
