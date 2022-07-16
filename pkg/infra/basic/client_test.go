//go:build !race
// +build !race

package basic_test

import (
	"context"

	"github.com/hyperledger-twgc/tape/e2e/mock"
	"github.com/hyperledger-twgc/tape/pkg/infra/basic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("Client", func() {

	Context("Client Error handling", func() {
		dummy := basic.Node{
			Addr: "invalid_addr",
		}
		logger := log.New()

		It("captures error from endorser", func() {
			_, err := basic.CreateEndorserClient(dummy, logger)
			Expect(err).Should(MatchError(ContainSubstring("error connecting to invalid_addr")))
		})
		It("captures error from broadcaster", func() {
			_, err := basic.CreateBroadcastClient(context.Background(), dummy, logger)
			Expect(err).Should(MatchError(ContainSubstring("error connecting to invalid_addr")))
		})
		It("captures error from DeliverFilter", func() {
			_, err := basic.CreateDeliverFilteredClient(context.Background(), dummy, logger)
			Expect(err).Should(MatchError(ContainSubstring("error connecting to invalid_addr")))
		})

		It("captures error from CreateDeliverClient", func() {
			_, err := basic.CreateDeliverClient(dummy)
			Expect(err).Should(MatchError(ContainSubstring("failed to create new connection")))
		})
	})

	Context("connect with mock peer", func() {
		var mockserver *mock.Server
		var peerNode, OrdererNode basic.Node
		logger := log.New()

		BeforeEach(func() {
			mockserver, _ = mock.NewServer(1, nil)
			peerNode = basic.Node{
				Addr: mockserver.PeersAddresses()[0],
			}
			OrdererNode = basic.Node{
				Addr: mockserver.OrderAddr(),
			}
			mockserver.Start()
		})

		AfterEach(func() {
			mockserver.Stop()
		})

		It("connect with mock endorsers", func() {
			_, err := basic.CreateEndorserClient(peerNode, logger)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("connect with mock broadcasters", func() {
			_, err := basic.CreateBroadcastClient(context.Background(), peerNode, logger)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("connect with mock DeliverFilter", func() {
			_, err := basic.CreateDeliverFilteredClient(context.Background(), OrdererNode, logger)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("connect with mock CreateDeliverClient", func() {
			_, err := basic.CreateDeliverClient(OrdererNode)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("wrong addr test", func() {
			dummy := basic.Node{
				Addr:          "invalid_addr",
				TLSCACertByte: []byte(""),
				TLSCAKey:      "123",
				TLSCARoot:     "234",
				TLSCARootByte: []byte(""),
			}
			_, err := basic.CreateGRPCClient(dummy)
			Expect(err).Should(HaveOccurred())
		})
	})

})
