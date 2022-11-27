package observer_test

import (
	"context"
	"sync"

	"github.com/hyperledger-twgc/tape/pkg/infra/observer"

	"github.com/google/uuid"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func newAddressedBlock(addr int, blockNum uint64) *observer.AddressedBlock {
	uuid, _ := uuid.NewRandom()
	FilteredTransactions := make([]*peer.FilteredTransaction, 0)
	FilteredTransactions = append(FilteredTransactions, &peer.FilteredTransaction{Txid: uuid.String()})
	data := &observer.AddressedBlock{Address: addr, FilteredBlock: &peer.FilteredBlock{Number: blockNum, FilteredTransactions: FilteredTransactions}}
	return data
}

var _ = Describe("BlockCollector", func() {

	var logger *log.Logger

	BeforeEach(func() {
		logger = log.New()

	})

	Context("Async Commit", func() {
		It("should work with threshold 1 and observer 1", func() {
			block := make(chan *observer.AddressedBlock)
			done := make(chan struct{})
			var once sync.Once
			instance, err := observer.NewBlockCollector(1, 1, context.Background(), block, done, 2, false, logger, &once, true)
			Expect(err).NotTo(HaveOccurred())

			go instance.Start()

			block <- newAddressedBlock(0, 0)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(0, 1)
			Eventually(done).Should(BeClosed())
		})

		It("should work with threshold 1 and observer 2", func() {
			block := make(chan *observer.AddressedBlock)
			done := make(chan struct{})
			var once sync.Once
			instance, err := observer.NewBlockCollector(1, 2, context.Background(), block, done, 2, false, logger, &once, true)
			Expect(err).NotTo(HaveOccurred())

			go instance.Start()

			block <- newAddressedBlock(0, 0)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(1, 0)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(0, 1)
			Eventually(done).Should(BeClosed())

			select {
			case block <- newAddressedBlock(1, 1):
			default:
				Fail("Block collector should still be able to consume blocks")
			}
		})

		It("should work with threshold 4 and observer 4", func() {
			block := make(chan *observer.AddressedBlock)
			done := make(chan struct{})
			var once sync.Once
			instance, err := observer.NewBlockCollector(4, 4, context.Background(), block, done, 2, false, logger, &once, true)
			Expect(err).NotTo(HaveOccurred())

			go instance.Start()

			block <- newAddressedBlock(0, 1)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(1, 1)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(2, 1)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(3, 1)
			Consistently(done).ShouldNot(BeClosed())

			block <- newAddressedBlock(0, 0)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(1, 0)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(2, 0)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(3, 0)
			Eventually(done).Should(BeClosed())
		})

		It("should work with threshold 2 and observer 4", func() {
			block := make(chan *observer.AddressedBlock)
			done := make(chan struct{})
			var once sync.Once
			instance, err := observer.NewBlockCollector(2, 4, context.Background(), block, done, 1, false, logger, &once, true)
			Expect(err).NotTo(HaveOccurred())

			go instance.Start()

			block <- newAddressedBlock(0, 0)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(1, 0)
			Eventually(done).Should(BeClosed())
		})

		PIt("should not count tx for repeated block", func() {
			block := make(chan *observer.AddressedBlock)
			done := make(chan struct{})
			var once sync.Once
			instance, err := observer.NewBlockCollector(1, 1, context.Background(), block, done, 2, false, logger, &once, true)
			Expect(err).NotTo(HaveOccurred())

			go instance.Start()

			block <- newAddressedBlock(0, 0)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(0, 0)
			Consistently(done).ShouldNot(BeClosed())

			block <- newAddressedBlock(0, 1)
			Eventually(done).Should(BeClosed())
		})

		It("should return err when threshold is greater than total", func() {
			block := make(chan *observer.AddressedBlock)
			done := make(chan struct{})
			var once sync.Once
			instance, err := observer.NewBlockCollector(2, 1, context.Background(), block, done, 2, false, logger, &once, true)
			Expect(err).Should(MatchError("threshold [2] must be less than or equal to total [1]"))
			Expect(instance).Should(BeNil())
		})

		It("should return err when threshold or total is zero", func() {
			block := make(chan *observer.AddressedBlock)
			done := make(chan struct{})
			var once sync.Once
			instance, err := observer.NewBlockCollector(0, 1, context.Background(), block, done, 2, false, logger, &once, true)
			Expect(err).Should(MatchError("threshold and total must be greater than zero"))
			Expect(instance).Should(BeNil())

			instance, err = observer.NewBlockCollector(1, 0, context.Background(), block, done, 2, false, logger, &once, true)
			Expect(err).Should(MatchError("threshold and total must be greater than zero"))
			Expect(instance).Should(BeNil())
		})

		It("Should supports parallel committers", func() {
			block := make(chan *observer.AddressedBlock)
			done := make(chan struct{})
			var once sync.Once
			instance, err := observer.NewBlockCollector(100, 100, context.Background(), block, done, 1, false, logger, &once, true)
			Expect(err).NotTo(HaveOccurred())

			go instance.Start()

			var wg sync.WaitGroup
			wg.Add(100)
			for i := 0; i < 100; i++ {
				go func(idx int) {
					defer wg.Done()
					block <- newAddressedBlock(idx, 0)
				}(i)
			}
			wg.Wait()
			Eventually(done).Should(BeClosed())
		})

		It("Should supports threshold 3 and observer 5 as parallel committers", func() {
			block := make(chan *observer.AddressedBlock)
			done := make(chan struct{})
			var once sync.Once
			instance, err := observer.NewBlockCollector(3, 5, context.Background(), block, done, 10, false, logger, &once, true)
			Expect(err).NotTo(HaveOccurred())

			go instance.Start()

			for i := 0; i < 3; i++ {
				go func(idx int) {
					for j := 0; j < 10; j++ {
						block <- newAddressedBlock(idx, uint64(j))
					}
				}(i)
			}
			Eventually(done).Should(BeClosed())
		})

		It("Should supports threshold 5 and observer 5 as parallel committers", func() {
			block := make(chan *observer.AddressedBlock)
			done := make(chan struct{})
			var once sync.Once
			instance, err := observer.NewBlockCollector(5, 5, context.Background(), block, done, 10, false, logger, &once, true)
			Expect(err).NotTo(HaveOccurred())

			go instance.Start()
			for i := 0; i < 5; i++ {
				go func(idx int) {
					for j := 0; j < 10; j++ {
						block <- newAddressedBlock(idx, uint64(j))
					}
				}(i)
			}
			Eventually(done).Should(BeClosed())
		})
	})
})
