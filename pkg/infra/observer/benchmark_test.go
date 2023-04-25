//go:build !race
// +build !race

package observer_test

import (
	"context"
	"sync"
	"testing"

	"github.com/opentracing/opentracing-go"

	"github.com/hyperledger-twgc/tape/e2e/mock"
	"github.com/hyperledger-twgc/tape/pkg/infra/basic"
	"github.com/hyperledger-twgc/tape/pkg/infra/observer"
	"github.com/hyperledger-twgc/tape/pkg/infra/trafficGenerator"

	"github.com/google/uuid"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	log "github.com/sirupsen/logrus"
)

func StartProposer(ctx context.Context, signed, processed chan *basic.Elements, logger *log.Logger, threshold int, addr string) {
	peer := basic.Node{
		Addr: addr,
	}
	tr, closer := basic.Init("test")
	defer closer.Close()
	opentracing.SetGlobalTracer(tr)
	rule := `
	package tape

default allow = false
		
allow {
  1 == 1
}
	`
	Proposer, _ := trafficGenerator.CreateProposer(peer, logger, rule)
	go Proposer.Start(ctx, signed, processed)
}

func benchmarkNPeer(concurrency int, b *testing.B) {
	processed := make(chan *basic.Elements, 10)
	signeds := make([]chan *basic.Elements, concurrency)
	ctx, cancel := context.WithCancel(context.Background())
	logger := log.New()
	defer cancel()
	for i := 0; i < concurrency; i++ {
		signeds[i] = make(chan *basic.Elements, 10)
		mockpeer, err := mock.NewServer(1, nil)
		if err != nil {
			b.Fatal(err)
		}
		mockpeer.Start()
		defer mockpeer.Stop()
		StartProposer(ctx, signeds[i], processed, logger, concurrency, mockpeer.PeersAddresses()[0])
	}
	b.ReportAllocs()
	b.ResetTimer()
	go func() {
		for i := 0; i < b.N; i++ {
			uuid, _ := uuid.NewRandom()
			span := opentracing.GlobalTracer().StartSpan("start transcation process", opentracing.Tag{Key: "txid", Value: uuid.String()})
			ed_span := opentracing.GlobalTracer().StartSpan("endorsement", opentracing.Tag{Key: "txid", Value: uuid.String()})
			data := &basic.Elements{SignedProp: &peer.SignedProposal{}, TxId: uuid.String(), Span: span, EndorsementSpan: ed_span}
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

func benchmarkAsyncCollector(concurrent int, b *testing.B) {
	block := make(chan *observer.AddressedBlock, 100)
	done := make(chan struct{})
	logger := log.New()

	var once sync.Once
	instance, _ := observer.NewBlockCollector(concurrent, concurrent, context.Background(), block, done, b.N, false, logger, &once, true)
	go instance.Start()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < concurrent; i++ {
		go func(idx int) {
			for j := 0; j < b.N; j++ {
				uuid, _ := uuid.NewRandom()
				FilteredTransactions := make([]*peer.FilteredTransaction, 0)
				FilteredTransactions = append(FilteredTransactions, &peer.FilteredTransaction{Txid: uuid.String()})
				data := &observer.AddressedBlock{Address: idx, FilteredBlock: &peer.FilteredBlock{Number: uint64(j), FilteredTransactions: FilteredTransactions}}
				block <- data
			}
		}(i)
	}
	<-done
	b.StopTimer()
}

func BenchmarkAsyncCollector1(b *testing.B)  { benchmarkAsyncCollector(1, b) }
func BenchmarkAsyncCollector2(b *testing.B)  { benchmarkAsyncCollector(2, b) }
func BenchmarkAsyncCollector4(b *testing.B)  { benchmarkAsyncCollector(4, b) }
func BenchmarkAsyncCollector8(b *testing.B)  { benchmarkAsyncCollector(8, b) }
func BenchmarkAsyncCollector16(b *testing.B) { benchmarkAsyncCollector(16, b) }
