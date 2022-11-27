//go:build !race
// +build !race

package trafficGenerator_test

import (
	"context"
	"time"

	"github.com/hyperledger-twgc/tape/e2e/mock"
	"github.com/hyperledger-twgc/tape/pkg/infra/basic"
	"github.com/hyperledger-twgc/tape/pkg/infra/trafficGenerator"
	"github.com/opentracing/opentracing-go"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("Proposer", func() {

	var addr string
	var logger = log.New()
	var processed chan *basic.Elements

	BeforeEach(func() {
		server, _ := mock.NewServer(1, nil)
		server.Start()
		addr = server.PeersAddresses()[0]
	})

	Context("CreateProposer", func() {
		It("successfully creates a proposer", func() {
			dummy := basic.Node{
				Addr: addr,
			}
			rule := "1 == 1"
			Proposer, err := trafficGenerator.CreateProposer(dummy, logger, rule)
			Expect(Proposer.Addr).To(Equal(addr))
			Expect(err).NotTo(HaveOccurred())
		})

		It("handle error ", func() {
			dummy := basic.Node{
				Addr: "invalid_addr",
			}
			rule := "1 == 1"
			_, err := trafficGenerator.CreateProposer(dummy, logger, rule)
			Expect(err).Should(MatchError(ContainSubstring("error connecting to invalid_addr")))
		})
	})

	Context("CreateBroadcasters", func() {
		It("successfully creates a broadcaster", func() {
			dummy := basic.Node{
				Addr: addr,
			}
			Broadcaster, err := trafficGenerator.CreateBroadcaster(context.Background(), dummy, logger)
			Expect(Broadcaster).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		It("captures connection errors", func() {
			dummy := basic.Node{
				Addr: "invalid_addr",
			}
			_, err := trafficGenerator.CreateBroadcaster(context.Background(), dummy, logger)
			Expect(err).Should(MatchError(ContainSubstring("error connecting to invalid_addr")))
		})
	})

	Context("Tape should do less for prepare and summary endorsement process", func() {
		// 0.002 here for mac testing on azp
		// For ginkgo,
		// You may only call Measure from within a Describe, Context or When
		// So here only tested with concurrent as 8 peers
		It("it should do endorsement efficiently for 2 peers", Serial, Label("measurement"), func() {
			experiment := gmeasure.NewExperiment("Tape Peers")
			AddReportEntry(experiment.Name, experiment)

			tracer, closer := basic.Init("test")
			defer closer.Close()
			opentracing.SetGlobalTracer(tracer)
			span := opentracing.GlobalTracer().StartSpan("start transcation process")
			defer span.Finish()
			peerNum := 2
			processed = make(chan *basic.Elements, 10)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			signeds := make([]chan *basic.Elements, peerNum)
			mockpeer, err := mock.NewServer(peerNum, nil)
			Expect(err).NotTo(HaveOccurred())
			mockpeer.Start()
			defer mockpeer.Stop()
			for i := 0; i < peerNum; i++ {
				signeds[i] = make(chan *basic.Elements, 10)
				StartProposer(ctx, signeds[i], processed, logger, peerNum, mockpeer.PeersAddresses()[i])
			}

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("process", func() {
					data := &basic.Elements{SignedProp: &peer.SignedProposal{}, TxId: "123", Span: span, EndorsementSpan: span}
					for _, s := range signeds {
						s <- data
					}
					<-processed
				})
			}, gmeasure.SamplingConfig{N: 100, Duration: time.Second})

			repaginationStats := experiment.GetStats("process")
			medianDuration := repaginationStats.DurationFor(gmeasure.StatMedian)

			Expect(medianDuration).To(BeNumerically("<", 2*time.Millisecond))
		})
	})
})

func StartProposer(ctx context.Context, signed, processed chan *basic.Elements, logger *log.Logger, threshold int, addr string) {
	peer := basic.Node{
		Addr: addr,
	}
	rule := `
	package tape

	default allow = true
	`
	Proposer, _ := trafficGenerator.CreateProposer(peer, logger, rule)
	go Proposer.Start(ctx, signed, processed)
}
