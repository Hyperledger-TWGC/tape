package trafficGenerator_test

import (
	"github.com/Hyperledger-TWGC/tape/pkg/infra/trafficGenerator"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PolicyHandler", func() {
	It("should pass", func() {
		input := make([]string, 2)
		input[0] = "org1"
		//input[1] = "org2"
		rs, err := trafficGenerator.Check(input)
		Expect(err).NotTo(HaveOccurred())
		Expect(rs).To(BeTrue())
	})

	It("should not pass", func() {
		input := make([]string, 2)
		rs, err := trafficGenerator.Check(input)
		Expect(err).NotTo(HaveOccurred())
		Expect(rs).To(BeFalse())
	})
})
