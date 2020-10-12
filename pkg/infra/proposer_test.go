package infra_test

import (
	"github.com/guoger/tape/pkg/infra"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("Proposer", func() {

	var addr string
	var logger = log.New()

	BeforeEach(func() {
		srv := &mocks.MockEndorserServer{}
		addr = srv.Start("127.0.0.1:0")
	})

	Context("CreateProposer", func() {
		It("successfully creates a proposer", func() {
			dummy := infra.Node{
				Addr: addr,
			}
			Proposer, err := infra.CreateProposer(dummy, logger)
			Expect(Proposer.Addr).To(Equal(addr))
			Expect(err).NotTo(HaveOccurred())
		})

		It("handle error ", func() {
			dummy := infra.Node{
				Addr: "invalid_addr",
			}
			_, err := infra.CreateProposer(dummy, logger)
			Expect(err).Should(MatchError(ContainSubstring("error connecting to invalid_addr")))
		})
	})

	Context("CreateBroadcasters", func() {
		It("successfully creates a broadcaster", func() {
			dummy := infra.Node{
				Addr: addr,
			}
			Broadcaster, err := infra.CreateBroadcaster(dummy, logger)
			Expect(Broadcaster).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		It("captures connection errors", func() {
			dummy := infra.Node{
				Addr: "invalid_addr",
			}
			_, err := infra.CreateBroadcaster(dummy, logger)
			Expect(err).Should(MatchError(ContainSubstring("error connecting to invalid_addr")))
		})
	})
})
