package bitmap_test

import (
	"github.com/hyperledger-twgc/tape/pkg/infra/bitmap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bitmap", func() {

	Context("New BitsMap", func() {
		It("the environment is properly set", func() {
			b, err := bitmap.NewBitMap(4)
			Expect(err).To(BeNil())
			Expect(b.Cap()).To(Equal(4))
			Expect(b.Count()).To(Equal(0))
			Expect(b.BitsLen()).To(Equal(1))

			b, err = bitmap.NewBitMap(65)
			Expect(err).To(BeNil())
			Expect(b.Cap()).To(Equal(65))
			Expect(b.Count()).To(Equal(0))
			Expect(b.BitsLen()).To(Equal(2))
		})

		It("should error which cap is less than 1", func() {
			_, err := bitmap.NewBitMap(0)
			Expect(err).NotTo(BeNil())

			_, err = bitmap.NewBitMap(-1)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Operate BitsMap", func() {
		It("the len of bits is just one ", func() {
			b, err := bitmap.NewBitMap(4)
			Expect(err).To(BeNil())
			b.Set(0)
			Expect(b.Count()).To(Equal(1))
			b.Set(2)
			Expect(b.Count()).To(Equal(2))
			ok := b.Has(0)
			Expect(ok).To(BeTrue())
			ok = b.Has(2)
			Expect(ok).To(BeTrue())
			ok = b.Has(1)
			Expect(ok).To(BeFalse())
			ok = b.Has(4)
			Expect(ok).To(BeFalse())

			b.Set(4)
			Expect(b.Count()).To(Equal(2))
			b.Set(2)
			Expect(b.Count()).To(Equal(2))
		})

		It("the len of bits is more than one", func() {
			b, err := bitmap.NewBitMap(80)
			Expect(err).To(BeNil())
			b.Set(0)
			Expect(b.Count()).To(Equal(1))
			b.Set(2)
			Expect(b.Count()).To(Equal(2))
			b.Set(70)
			Expect(b.Count()).To(Equal(3))
			b.Set(79)
			Expect(b.Count()).To(Equal(4))
			ok := b.Has(0)
			Expect(ok).To(BeTrue())
			ok = b.Has(2)
			Expect(ok).To(BeTrue())
			ok = b.Has(70)
			Expect(ok).To(BeTrue())
			ok = b.Has(79)
			Expect(ok).To(BeTrue())
			ok = b.Has(1)
			Expect(ok).To(BeFalse())
			ok = b.Has(3)
			Expect(ok).To(BeFalse())
			ok = b.Has(69)
			Expect(ok).To(BeFalse())

			b.Set(80)
			Expect(b.Count()).To(Equal(4))
			b.Set(2)
			Expect(b.Count()).To(Equal(4))
		})
	})
})
