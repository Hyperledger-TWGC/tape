package infra_test

import (
	"github.com/guoger/stupid/pkg/infra"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Proposer", func() {

	var addr string

	BeforeEach(func() {
		srv := &mocks.MockEndorserServer{}
		addr = srv.Start("127.0.0.1:0")
	})

	Context("CreateProposer", func() {
		It("successfully creates a proposer", func() {
			Proposer := infra.CreateProposer(addr, &infra.Crypto{}, nil)
			Expect(Proposer.Addr).To(Equal(addr))
		})

		It("error happen creates a proposer", func() {
			Expect(func() {
				infra.CreateProposer("invalid_addr", &infra.Crypto{}, nil)
			}).To(Panic())
		})
	})
})
