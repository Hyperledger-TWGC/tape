package infra_test

import (
	"github.com/guoger/stupid/pkg/infra"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {

	Context("Should Error handle", func() {
		dummy := infra.Node{
			Addr: "invalid_addr",
		}
		It("for endorser", func() {
			_, err := infra.CreateEndorserClient(dummy)
			Expect(err).Should(MatchError(ContainSubstring("error connect to invalid_addr")))
		})
		It("for broadcaster", func() {
			_, err := infra.CreateBroadcastClient(dummy)
			Expect(err).Should(MatchError(ContainSubstring("error connect to invalid_addr")))
		})
		It("for DeliverFilter", func() {
			_, err := infra.CreateDeliverFilteredClient(dummy)
			Expect(err).Should(MatchError(ContainSubstring("error connect to invalid_addr")))
		})
	})

})
