package infra

import (
	"testing"
	"time"

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

func benchmarkNPeer(concurrency int, b *testing.B) {
	processed := make(chan *Elements, 10)
	done := make(chan struct{})
	defer close(done)
	signeds := make([]chan *Elements, concurrency)
	for i := 0; i < concurrency; i++ {
		signeds[i] = make(chan *Elements, 10)
		mockpeer, err := mock.NewServer(1, nil)
		if err != nil {
			b.Fatal(err)
		}
		mockpeer.Start()
		defer mockpeer.Stop()
		StartProposer(signeds[i], processed, done, nil, concurrency, mockpeer.PeersAddresses()[0])
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

func benchmarkSyncCollector(concurrency int, b *testing.B) {
	instance, _ := NewBlockCollector(concurrency, concurrency)
	processed := make(chan struct{}, b.N)
	defer close(processed)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < concurrency; i++ {
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

func BenchmarkSyncCollector1(b *testing.B)  { benchmarkSyncCollector(1, b) }
func BenchmarkSyncCollector2(b *testing.B)  { benchmarkSyncCollector(2, b) }
func BenchmarkSyncCollector4(b *testing.B)  { benchmarkSyncCollector(4, b) }
func BenchmarkSyncCollector8(b *testing.B)  { benchmarkSyncCollector(8, b) }
func BenchmarkSyncCollector16(b *testing.B) { benchmarkSyncCollector(16, b) }

func benchmarkAsyncCollector(concurrent int, b *testing.B) {
	instance, _ := NewBlockCollector(concurrent, concurrent)
	block := make(chan *peer.FilteredBlock, 100)
	done := make(chan struct{})
	go instance.Start(block, done, b.N, time.Now(), false)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < concurrent; i++ {
		go func() {
			for j := 0; j < b.N; j++ {
				block <- &peer.FilteredBlock{
					Number:               uint64(j),
					FilteredTransactions: make([]*peer.FilteredTransaction, 1),
				}
			}
		}()
	}
	<-done
	b.StopTimer()
}

func BenchmarkAsyncCollector1(b *testing.B)  { benchmarkAsyncCollector(1, b) }
func BenchmarkAsyncCollector2(b *testing.B)  { benchmarkAsyncCollector(2, b) }
func BenchmarkAsyncCollector4(b *testing.B)  { benchmarkAsyncCollector(4, b) }
func BenchmarkAsyncCollector8(b *testing.B)  { benchmarkAsyncCollector(8, b) }
func BenchmarkAsyncCollector16(b *testing.B) { benchmarkAsyncCollector(16, b) }
