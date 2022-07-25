package trafficGenerator_test

import (
	"regexp"
	"strconv"

	"github.com/hyperledger-twgc/tape/pkg/infra/trafficGenerator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Proposal", func() {
	Context("ConvertString", func() {
		It("work accordingly for string", func() {
			input := "data"
			data, err := trafficGenerator.ConvertString(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal("data"))
		})

		It("work accordingly for random str", func() {
			input := "randomString1"
			data, err := trafficGenerator.ConvertString(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(data)).To(Equal(1))
		})

		It("work accordingly for uuid", func() {
			input := "uuid"
			data, err := trafficGenerator.ConvertString(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(data) > 0).To(BeTrue())
		})

		It("work accordingly for randomNumber", func() {
			input := "randomNumber1_9"
			data, err := trafficGenerator.ConvertString(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(data)).To(Equal(1))
			num, err := strconv.Atoi(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(num > 0).To(BeTrue())
		})

		It("work accordingly for string mix mode", func() {
			input := "{\"key\":\"randomNumber1_50\",\"key1\":\"randomNumber1_20\"}"
			data, err := trafficGenerator.ConvertString(input)
			Expect(err).NotTo(HaveOccurred())
			regString, err := regexp.Compile("{\"key\":\"\\d*\",\"key1\":\"\\d*\"}")
			Expect(err).NotTo(HaveOccurred())
			Expect(regString.MatchString(data)).To(BeTrue())
		})

		It("work accordingly for string mix mode 2", func() {
			input := "{\"k1\":\"uuid\",\"key2\":\"randomNumber10000_20000\",\"keys\":\"randomString10\"}"
			data, err := trafficGenerator.ConvertString(input)
			Expect(err).NotTo(HaveOccurred())
			regString, err := regexp.Compile("{\"k1\":\".*\",\"key2\":\"\\d*\",\"keys\":\".*\"}")
			Expect(err).NotTo(HaveOccurred())
			Expect(regString.MatchString(data)).To(BeTrue())
		})

		It("handle edge case for randmon number", func() {
			input := "randomNumber1_00"
			_, err := trafficGenerator.ConvertString(input)
			Expect(err).To(HaveOccurred())
		})

		It("handle edge case for randmon number", func() {
			input := "randomNumber1_1"
			_, err := trafficGenerator.ConvertString(input)
			Expect(err).To(HaveOccurred())
		})

		It("handle edge case for randomstring", func() {
			input := "0randomString166666666011010"
			_, err := trafficGenerator.ConvertString(input)
			Expect(err).To(HaveOccurred())
		})
	})
})
