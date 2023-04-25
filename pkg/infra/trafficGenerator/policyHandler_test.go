package trafficGenerator_test

import (
	"github.com/hyperledger-twgc/tape/pkg/infra/basic"
	"github.com/hyperledger-twgc/tape/pkg/infra/trafficGenerator"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const org1 = "org1"
const org2 = "org2"

var _ = Describe("PolicyHandler", func() {
	It("should pass when org1 with rule org1 or org2", func() {
		input := make([]string, 2)
		input[0] = org1

		//input[1] = "org2"
		rule := `package tape

		default allow = false
		allow {
			input[_] == "` + org1 + `"
		}
		
		allow {
			input[_] == "` + org2 + `"
		}`

		instance := &basic.Elements{
			Orgs: input,
		}
		rs, err := trafficGenerator.CheckPolicy(instance, rule)
		Expect(err).NotTo(HaveOccurred())
		Expect(rs).To(BeTrue())
	})

	It("should not pass when null with rule org1 or org2", func() {
		input := make([]string, 2)
		rule := `package tape

		default allow = false
		allow {
			input[_] == "` + org1 + `"
		}
		
		allow {
			input[_] == "` + org2 + `"
		}`
		instance := &basic.Elements{
			Orgs: input,
		}
		rs, err := trafficGenerator.CheckPolicy(instance, rule)
		Expect(err).NotTo(HaveOccurred())
		Expect(rs).To(BeFalse())
	})

	It("should not pass when org1 with rule org1 and org2", func() {
		input := make([]string, 2)
		input[0] = "org1"
		rule := `package tape

		default allow = false
		allow {
			input[_] == "` + org1 + `"
			input[_] == "` + org2 + `"
		}`
		instance := &basic.Elements{
			Orgs: input,
		}
		rs, err := trafficGenerator.CheckPolicy(instance, rule)
		Expect(err).NotTo(HaveOccurred())
		Expect(rs).To(BeFalse())
	})

	It("should pass with rule org1 and org2", func() {
		input := make([]string, 2)
		input[0] = "org1"
		input[1] = "org2"
		rule := `package tape

		default allow = false
		allow {
			input[_] == "` + org1 + `"
			input[_] == "` + org2 + `"
		}
		`
		instance := &basic.Elements{
			Orgs: input,
		}
		rs, err := trafficGenerator.CheckPolicy(instance, rule)
		Expect(err).NotTo(HaveOccurred())
		Expect(rs).To(BeTrue())
	})

	It("should pass with rule org1 and org2", func() {
		input := make([]string, 2)
		input[0] = org1
		input[1] = org2
		rule := `package tape

		default allow = false
		allow {
			1 == 1
		}
		`
		instance := &basic.Elements{
			Orgs: input,
		}
		rs, err := trafficGenerator.CheckPolicy(instance, rule)
		Expect(err).NotTo(HaveOccurred())
		Expect(rs).To(BeTrue())
	})

	It("Same instance can't pass twice", func() {
		input := make([]string, 2)
		input[0] = org1
		input[1] = org2
		rule := `package tape

		default allow = false
		allow {
			input[_] == "` + org1 + `"
			input[_] == "` + org2 + `"
		}
		`
		instance := &basic.Elements{
			Orgs: input,
		}
		rs, err := trafficGenerator.CheckPolicy(instance, rule)
		Expect(err).NotTo(HaveOccurred())
		Expect(rs).To(BeTrue())

		rs, err = trafficGenerator.CheckPolicy(instance, rule)
		Expect(err).NotTo(HaveOccurred())
		Expect(rs).To(BeFalse())
	})
})
