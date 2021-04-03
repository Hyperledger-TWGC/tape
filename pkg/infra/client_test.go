package infra_test

import (
	"tape/pkg/infra"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("Client", func() {

	Context("Client Error handling", func() {
		dummy := infra.Node{
			Addr: "invalid_addr",
		}
		logger := log.New()

		It("captures error from endorser", func() {
			_, err := infra.CreateEndorserClient(dummy, logger)
			Expect(err).Should(MatchError(ContainSubstring("error connecting to invalid_addr")))
		})
		It("captures error from broadcaster", func() {
			ctx, _ := infra.TapeContext()
			_, err := infra.CreateBroadcastClient(dummy, logger, ctx)
			Expect(err).Should(MatchError(ContainSubstring("error connecting to invalid_addr")))
		})
		It("captures error from DeliverFilter", func() {
			ctx, _ := infra.TapeContext()
			_, err := infra.CreateDeliverFilteredClient(dummy, logger, ctx)
			Expect(err).Should(MatchError(ContainSubstring("error connecting to invalid_addr")))
		})
	})

})
