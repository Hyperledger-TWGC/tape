package infra_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"tape/pkg/infra"
)

var _ = Describe("Client", func() {

	Context("Client Error handling", func() {
		dummy := infra.Node{
			Addr: "invalid_addr",
		}
		It("captures error from endorser", func() {
			_, err := infra.CreateEndorserClient(dummy)
			Expect(err).Should(MatchError(ContainSubstring("error connecting to invalid_addr")))
		})
		It("captures error from broadcaster", func() {
			_, err := infra.CreateBroadcastClient(dummy)
			Expect(err).Should(MatchError(ContainSubstring("error connecting to invalid_addr")))
		})
		It("captures error from DeliverFilter", func() {
			_, err := infra.CreateDeliverFilteredClient(dummy)
			Expect(err).Should(MatchError(ContainSubstring("error connecting to invalid_addr")))
		})
	})

})
