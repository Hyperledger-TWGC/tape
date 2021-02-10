package infra_test

import (
	"context"
	"sync"
	"tape/pkg/infra"
	"time"

	"github.com/hyperledger/fabric-protos-go/peer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BlockCollector", func() {

	now := time.Now()

	Context("Async Commit", func() {
		It("should work with threshold 1 and observer 1", func() {
			instance, err := infra.NewBlockCollector(1, 1)
			Expect(err).NotTo(HaveOccurred())

			block := make(chan *peer.FilteredBlock)
			done := make(chan struct{})
			go instance.Start(context.Background(), block, done, 2, time.Now(), false)

			block <- &peer.FilteredBlock{Number: 0, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Consistently(done).ShouldNot(BeClosed())
			block <- &peer.FilteredBlock{Number: 1, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Eventually(done).Should(BeClosed())
		})

		It("should work with threshold 1 and observer 2", func() {
			instance, err := infra.NewBlockCollector(1, 2)
			Expect(err).NotTo(HaveOccurred())

			block := make(chan *peer.FilteredBlock)
			done := make(chan struct{})
			go instance.Start(context.Background(), block, done, 2, time.Now(), false)

			block <- &peer.FilteredBlock{Number: 0, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Consistently(done).ShouldNot(BeClosed())
			block <- &peer.FilteredBlock{Number: 0, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Consistently(done).ShouldNot(BeClosed())
			block <- &peer.FilteredBlock{Number: 1, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Eventually(done).Should(BeClosed())

			select {
			case block <- &peer.FilteredBlock{Number: 1, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}:
			default:
				Fail("Block collector should still be able to consume blocks")
			}
		})

		It("should work with threshold 4 and observer 4", func() {
			instance, err := infra.NewBlockCollector(4, 4)
			Expect(err).NotTo(HaveOccurred())

			block := make(chan *peer.FilteredBlock)
			done := make(chan struct{})
			go instance.Start(context.Background(), block, done, 2, time.Now(), false)

			block <- &peer.FilteredBlock{Number: 1, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Consistently(done).ShouldNot(BeClosed())
			block <- &peer.FilteredBlock{Number: 1, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Consistently(done).ShouldNot(BeClosed())
			block <- &peer.FilteredBlock{Number: 1, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Consistently(done).ShouldNot(BeClosed())
			block <- &peer.FilteredBlock{Number: 1, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Consistently(done).ShouldNot(BeClosed())

			block <- &peer.FilteredBlock{Number: 0, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Consistently(done).ShouldNot(BeClosed())
			block <- &peer.FilteredBlock{Number: 0, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Consistently(done).ShouldNot(BeClosed())
			block <- &peer.FilteredBlock{Number: 0, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Consistently(done).ShouldNot(BeClosed())
			block <- &peer.FilteredBlock{Number: 0, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Eventually(done).Should(BeClosed())
		})

		It("should work with threshold 2 and observer 4", func() {
			instance, err := infra.NewBlockCollector(2, 4)
			Expect(err).NotTo(HaveOccurred())

			block := make(chan *peer.FilteredBlock)
			done := make(chan struct{})
			go instance.Start(context.Background(), block, done, 1, time.Now(), false)

			block <- &peer.FilteredBlock{FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Consistently(done).ShouldNot(BeClosed())
			block <- &peer.FilteredBlock{FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Eventually(done).Should(BeClosed())
		})

		PIt("should not count tx for repeated block", func() {
			instance, err := infra.NewBlockCollector(1, 1)
			Expect(err).NotTo(HaveOccurred())

			block := make(chan *peer.FilteredBlock)
			done := make(chan struct{})
			go instance.Start(context.Background(), block, done, 2, time.Now(), false)

			block <- &peer.FilteredBlock{Number: 0, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Consistently(done).ShouldNot(BeClosed())
			block <- &peer.FilteredBlock{Number: 0, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Consistently(done).ShouldNot(BeClosed())

			block <- &peer.FilteredBlock{Number: 1, FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
			Eventually(done).Should(BeClosed())
		})

		It("should return err when threshold is greater than total", func() {
			instance, err := infra.NewBlockCollector(2, 1)
			Expect(err).Should(MatchError("threshold [2] must be less than or equal to total [1]"))
			Expect(instance).Should(BeNil())
		})

		It("Should supports parallel committers", func() {
			instance, err := infra.NewBlockCollector(100, 100)
			Expect(err).NotTo(HaveOccurred())

			block := make(chan *peer.FilteredBlock)
			done := make(chan struct{})
			go instance.Start(context.Background(), block, done, 1, time.Now(), false)

			var wg sync.WaitGroup
			wg.Add(100)
			for i := 0; i < 100; i++ {
				go func() {
					defer wg.Done()
					block <- &peer.FilteredBlock{FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
				}()
			}
			wg.Wait()
			Eventually(done).Should(BeClosed())
		})

		It("Should supports threshold 3 and observer 5 as parallel committers", func() {
			instance, err := infra.NewBlockCollector(3, 5)
			Expect(err).NotTo(HaveOccurred())

			block := make(chan *peer.FilteredBlock)
			done := make(chan struct{})
			go instance.Start(context.Background(), block, done, 10, time.Now(), false)

			for i := 0; i < 3; i++ {
				go func() {
					for j := 0; j < 10; j++ {
						block <- &peer.FilteredBlock{
							Number:               uint64(j),
							FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
					}
				}()
			}
			Eventually(done).Should(BeClosed())
		})

		It("Should supports threshold 5 and observer 5 as parallel committers", func() {
			instance, err := infra.NewBlockCollector(5, 5)
			Expect(err).NotTo(HaveOccurred())

			block := make(chan *peer.FilteredBlock)
			done := make(chan struct{})

			go instance.Start(context.Background(), block, done, 10, time.Now(), false)
			for i := 0; i < 5; i++ {
				go func() {
					for j := 0; j < 10; j++ {
						block <- &peer.FilteredBlock{
							Number:               uint64(j),
							FilteredTransactions: make([]*peer.FilteredTransaction, 1)}
					}
				}()
			}
			Eventually(done).Should(BeClosed())
		})
	})

	Context("Sync Commit", func() {
		It("should work with threshold 1 and observer 1", func() {
			finishCh := make(chan struct{})
			instance, err := infra.NewBlockCollector(1, 1)
			Expect(err).NotTo(HaveOccurred())
			ft := make([]*peer.FilteredTransaction, 1)
			fb := &peer.FilteredBlock{
				Number:               uint64(1),
				FilteredTransactions: ft,
			}
			block := &peer.DeliverResponse_FilteredBlock{
				FilteredBlock: fb,
			}
			Expect(instance.Commit(block, finishCh, now)).To(BeTrue())
		})

		It("should work with threshold 1 and observer 2", func() {
			finishCh := make(chan struct{})
			instance, err := infra.NewBlockCollector(1, 2)
			Expect(err).NotTo(HaveOccurred())
			ft := make([]*peer.FilteredTransaction, 1)
			fb := &peer.FilteredBlock{
				Number:               uint64(1),
				FilteredTransactions: ft,
			}
			block := &peer.DeliverResponse_FilteredBlock{
				FilteredBlock: fb,
			}
			Expect(instance.Commit(block, finishCh, now)).To(BeTrue())
			Expect(instance.Commit(block, finishCh, now)).To(BeFalse())
		})

		It("should work with threshold 4 and observer 4", func() {
			finishCh := make(chan struct{})
			instance, err := infra.NewBlockCollector(4, 4)
			Expect(err).NotTo(HaveOccurred())
			ft := make([]*peer.FilteredTransaction, 1)
			fb := &peer.FilteredBlock{
				Number:               uint64(1),
				FilteredTransactions: ft,
			}
			block := &peer.DeliverResponse_FilteredBlock{
				FilteredBlock: fb,
			}
			Expect(instance.Commit(block, finishCh, now)).To(BeFalse())
			Expect(instance.Commit(block, finishCh, now)).To(BeFalse())
			Expect(instance.Commit(block, finishCh, now)).To(BeFalse())
			Expect(instance.Commit(block, finishCh, now)).To(BeTrue())
		})

		It("should work with threshold 2 and observer 4", func() {
			finishCh := make(chan struct{})
			instance, err := infra.NewBlockCollector(2, 4)
			Expect(err).NotTo(HaveOccurred())
			ft := make([]*peer.FilteredTransaction, 1)
			fb := &peer.FilteredBlock{
				Number:               uint64(1),
				FilteredTransactions: ft,
			}
			block := &peer.DeliverResponse_FilteredBlock{
				FilteredBlock: fb,
			}
			Expect(instance.Commit(block, finishCh, now)).To(BeFalse())
			Expect(instance.Commit(block, finishCh, now)).To(BeTrue())
			Expect(instance.Commit(block, finishCh, now)).To(BeFalse())
			Expect(instance.Commit(block, finishCh, now)).To(BeFalse())
		})

		It("should return err when threshold is greater than total", func() {
			instance, err := infra.NewBlockCollector(2, 1)
			Expect(err).Should(MatchError("threshold [2] must be less than or equal to total [1]"))
			Expect(instance).Should(BeNil())
		})

		It("Should work with threshold 3 and observer 5 in parallel", func() {
			instance, _ := infra.NewBlockCollector(3, 5)
			finishCh := make(chan struct{})
			var wg sync.WaitGroup
			wg.Add(3)
			for i := 0; i < 3; i++ {
				go func() {
					defer wg.Done()
					ft := make([]*peer.FilteredTransaction, 1)
					fb := &peer.FilteredBlock{
						Number:               uint64(1),
						FilteredTransactions: ft,
					}
					block := &peer.DeliverResponse_FilteredBlock{
						FilteredBlock: fb,
					}
					if instance.Commit(block, finishCh, now) {
						close(finishCh)
					}
				}()
			}
			wg.Wait()
			Eventually(finishCh).Should(BeClosed())
			Expect(finishCh).To(BeClosed())
		})

		It("Should work with threshold 5 and observer 5 in parallel", func() {
			instance, _ := infra.NewBlockCollector(5, 5)
			finishCh := make(chan struct{})
			var wg sync.WaitGroup
			wg.Add(5)
			for i := 0; i < 5; i++ {
				go func() {
					defer wg.Done()
					ft := make([]*peer.FilteredTransaction, 1)
					fb := &peer.FilteredBlock{
						Number:               uint64(1),
						FilteredTransactions: ft,
					}
					block := &peer.DeliverResponse_FilteredBlock{
						FilteredBlock: fb,
					}
					if instance.Commit(block, finishCh, now) {
						close(finishCh)
					}
				}()
			}
			wg.Wait()
			Eventually(finishCh).Should(BeClosed())
			Expect(finishCh).To(BeClosed())
		})
	})
})
