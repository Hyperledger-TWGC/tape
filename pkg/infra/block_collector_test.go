package infra_test

import (
	"sync"
	"tape/pkg/infra"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BlockCollector", func() {

	Context("Commit", func() {
		It("should work with threshold 1 and observer 1", func() {
			instance, err := infra.NewBlockCollector(1, 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance.Commit(1)).To(BeTrue())
		})

		It("should work with threshold 1 and observer 2", func() {
			instance, err := infra.NewBlockCollector(1, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance.Commit(1)).To(BeTrue())
			Expect(instance.Commit(1)).To(BeFalse())
		})

		It("should work with threshold 4 and observer 4", func() {
			instance, err := infra.NewBlockCollector(4, 4)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance.Commit(1)).To(BeFalse())
			Expect(instance.Commit(1)).To(BeFalse())
			Expect(instance.Commit(1)).To(BeFalse())
			Expect(instance.Commit(1)).To(BeTrue())
		})

		It("should work with threshold 2 and observer 4", func() {
			instance, err := infra.NewBlockCollector(2, 4)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance.Commit(1)).To(BeFalse())
			Expect(instance.Commit(1)).To(BeTrue())
			Expect(instance.Commit(1)).To(BeFalse())
			Expect(instance.Commit(1)).To(BeFalse())
		})

		It("should return err when threshold is greater than total", func() {
			instance, err := infra.NewBlockCollector(2, 1)
			Expect(err).Should(MatchError("threshold [2] must be less than or equal to total [1]"))
			Expect(instance).Should(BeNil())
		})

		It("Should supports parallel committers", func() {
			instance, _ := infra.NewBlockCollector(100, 100)
			signal := make(chan struct{})
			var wg sync.WaitGroup
			wg.Add(100)
			for i := 0; i < 100; i++ {
				go func() {
					defer wg.Done()
					if instance.Commit(1) {
						close(signal)
					}
				}()
			}
			wg.Wait()
			Expect(signal).To(BeClosed())
		})
	})
})
