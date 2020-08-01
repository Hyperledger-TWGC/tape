package infra_test

import (
	"sync"
	"tape/pkg/infra"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BlockCollector", func() {

	Context("Commit", func() {
		It("should works when threshold 1 observer 1", func() {
			instance, err := infra.NewBlockCollector(1, 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance.Commit(uint64(1))).To(BeTrue())
		})

		It("should works when threshold 1 observer 2", func() {
			instance, err := infra.NewBlockCollector(1, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance.Commit(uint64(1))).To(BeTrue())
			Expect(instance.Commit(uint64(1))).To(BeFalse())
		})

		It("should works when threshold 4 observer 4", func() {
			instance, err := infra.NewBlockCollector(4, 4)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance.Commit(uint64(1))).To(BeFalse())
			Expect(instance.Commit(uint64(1))).To(BeFalse())
			Expect(instance.Commit(uint64(1))).To(BeFalse())
			Expect(instance.Commit(uint64(1))).To(BeTrue())
		})

		It("should works when threshold 2 observer 4", func() {
			instance, err := infra.NewBlockCollector(2, 4)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance.Commit(uint64(1))).To(BeFalse())
			Expect(instance.Commit(uint64(1))).To(BeTrue())
			Expect(instance.Commit(uint64(1))).To(BeFalse())
			Expect(instance.Commit(uint64(1))).To(BeFalse())
		})

		It("Should return err when threshold is bigger than observer", func() {
			instance, err := infra.NewBlockCollector(2, 1)
			Expect(err).Should(MatchError(ContainSubstring("commitThreshold should not bigger than committers, please check your config")))
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
					if instance.Commit(uint64(1)) {
						close(signal)
					}
				}()
			}
			wg.Wait()
			Expect(signal).To(BeClosed())
		})
	})
})
