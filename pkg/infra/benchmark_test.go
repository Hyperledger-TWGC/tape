package infra

import (
	"testing"

	"tape/e2e/mock"

	"github.com/hyperledger/fabric-protos-go/peer"
	log "github.com/sirupsen/logrus"
)

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
		mockpeer, err := mock.NewServer(1, nil)
		if err != nil {
			b.Fatal(err)
		}
		mockpeer.Start()
		defer mockpeer.Stop()
		StartProposer(signeds[i], processed, done, nil, concurrent, mockpeer.PeersAddresses()[0])
	}
	b.ReportAllocs()
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

func benchmarkCollector(concurrent int, b *testing.B) {
	instance, _ := NewBlockCollector(concurrent, concurrent)
	processed := make(chan struct{}, b.N)
	defer close(processed)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < concurrent; i++ {
		go func() {
			for j := 0; j < b.N; j++ {
				if instance.Commit(uint64(j)) {
					processed <- struct{}{}
				}
			}
		}()
	}
	var n int
	for n < b.N {
		<-processed
		n++
	}
	b.StopTimer()
}

func BenchmarkCollector1(b *testing.B) { benchmarkCollector(1, b) }
func BenchmarkCollector2(b *testing.B) { benchmarkCollector(2, b) }
func BenchmarkCollector4(b *testing.B) { benchmarkCollector(4, b) }
func BenchmarkCollector8(b *testing.B) { benchmarkCollector(8, b) }
