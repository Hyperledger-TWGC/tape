package infra_test

import (
	"context"
	"io/ioutil"
	"os"
	"sync"
	"tape/pkg/infra"
	"time"

	"github.com/hyperledger/fabric-protos-go/peer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func newAddressedBlock(idx int, blockNum uint64, isValidTx bool) *infra.AddressedBlock {
	tx := make([]*peer.FilteredTransaction, 1)
	if isValidTx {
		tx[0] = &peer.FilteredTransaction{TxValidationCode: peer.TxValidationCode_VALID}
	} else {
		tx[0] = &peer.FilteredTransaction{TxValidationCode: peer.TxValidationCode_NIL_ENVELOPE}
	}

	return &infra.AddressedBlock{PeerIdx: idx, FilteredBlock: &peer.FilteredBlock{Number: blockNum,
		FilteredTransactions: tx}}
}

var _ = Describe("BlockCollector", func() {

	Context("Async Commit", func() {
		It("should work with threshold 1 and observer 1", func() {
			instance, err := infra.NewBlockCollector(1, 1)
			Expect(err).NotTo(HaveOccurred())

			block := make(chan *infra.AddressedBlock)
			successRateBlock := make(chan *infra.AddressedBlock)
			done := make(chan struct{})
			go infra.CalSuccessRate(1, 2, successRateBlock)
			go instance.Start(context.Background(), block, successRateBlock, done, 2, time.Now(), false)

			block <- newAddressedBlock(0, 0, true)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(0, 1, true)
			Eventually(done).Should(BeClosed())
		})

		It("should work with threshold 1 and observer 2", func() {
			instance, err := infra.NewBlockCollector(1, 2)
			Expect(err).NotTo(HaveOccurred())

			block := make(chan *infra.AddressedBlock)
			successRateBlock := make(chan *infra.AddressedBlock)
			done := make(chan struct{})
			go infra.CalSuccessRate(2, 2, successRateBlock)
			go instance.Start(context.Background(), block, successRateBlock, done, 2, time.Now(), false)

			block <- newAddressedBlock(0, 0, true)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(1, 0, true)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(0, 1, true)
			Eventually(done).Should(BeClosed())

			select {
			case block <- newAddressedBlock(1, 1, true):
			default:
				Fail("Block collector should still be able to consume blocks")
			}
		})

		It("should work with threshold 4 and observer 4", func() {
			instance, err := infra.NewBlockCollector(4, 4)
			Expect(err).NotTo(HaveOccurred())

			block := make(chan *infra.AddressedBlock)
			successRateBlock := make(chan *infra.AddressedBlock)
			done := make(chan struct{})
			go infra.CalSuccessRate(4, 2, successRateBlock)
			go instance.Start(context.Background(), block, successRateBlock, done, 2, time.Now(), false)

			block <- newAddressedBlock(0, 1, true)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(1, 1, true)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(2, 1, true)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(3, 1, true)
			Consistently(done).ShouldNot(BeClosed())

			block <- newAddressedBlock(0, 0, true)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(1, 0, true)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(2, 0, true)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(3, 0, true)
			Eventually(done).Should(BeClosed())
		})

		It("should work with threshold 2 and observer 4", func() {
			instance, err := infra.NewBlockCollector(2, 4)
			Expect(err).NotTo(HaveOccurred())

			block := make(chan *infra.AddressedBlock)
			successRateBlock := make(chan *infra.AddressedBlock)
			done := make(chan struct{})
			go infra.CalSuccessRate(4, 1, successRateBlock)
			go instance.Start(context.Background(), block, successRateBlock, done, 1, time.Now(), false)

			block <- newAddressedBlock(0, 0, true)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(1, 0, true)
			Eventually(done).Should(BeClosed())
		})

		PIt("should not count tx for repeated block", func() {
			instance, err := infra.NewBlockCollector(1, 1)
			Expect(err).NotTo(HaveOccurred())

			block := make(chan *infra.AddressedBlock)
			successRateBlock := make(chan *infra.AddressedBlock)
			done := make(chan struct{})
			go infra.CalSuccessRate(1, 2, successRateBlock)
			go instance.Start(context.Background(), block, successRateBlock, done, 2, time.Now(), false)

			block <- newAddressedBlock(0, 0, true)
			Consistently(done).ShouldNot(BeClosed())
			block <- newAddressedBlock(0, 0, true)
			Consistently(done).ShouldNot(BeClosed())

			block <- newAddressedBlock(0, 1, true)
			Eventually(done).Should(BeClosed())
		})

		It("should return err when threshold is greater than total", func() {
			instance, err := infra.NewBlockCollector(2, 1)
			Expect(err).Should(MatchError("threshold [2] must be less than or equal to total [1]"))
			Expect(instance).Should(BeNil())
		})

		It("should return err when threshold or total is zero", func() {
			instance, err := infra.NewBlockCollector(0, 1)
			Expect(err).Should(MatchError("threshold and total must be greater than zero"))
			Expect(instance).Should(BeNil())

			instance, err = infra.NewBlockCollector(1, 0)
			Expect(err).Should(MatchError("threshold and total must be greater than zero"))
			Expect(instance).Should(BeNil())
		})

		It("Should supports parallel committers", func() {
			instance, err := infra.NewBlockCollector(100, 100)
			Expect(err).NotTo(HaveOccurred())

			block := make(chan *infra.AddressedBlock)
			successRateBlock := make(chan *infra.AddressedBlock)
			done := make(chan struct{})
			go infra.CalSuccessRate(100, 1, successRateBlock)
			go instance.Start(context.Background(), block, successRateBlock, done, 1, time.Now(), false)

			var wg sync.WaitGroup
			wg.Add(100)
			for i := 0; i < 100; i++ {
				go func(idx int) {
					defer wg.Done()
					block <- newAddressedBlock(idx, 0, true)
				}(i)
			}
			wg.Wait()
			Eventually(done).Should(BeClosed())
		})

		It("Should supports threshold 3 and observer 5 as parallel committers", func() {
			instance, err := infra.NewBlockCollector(3, 5)
			Expect(err).NotTo(HaveOccurred())

			block := make(chan *infra.AddressedBlock)
			successRateBlock := make(chan *infra.AddressedBlock)
			done := make(chan struct{})
			go infra.CalSuccessRate(5, 10, successRateBlock)
			go instance.Start(context.Background(), block, successRateBlock, done, 10, time.Now(), false)

			for i := 0; i < 3; i++ {
				go func(idx int) {
					for j := 0; j < 10; j++ {
						block <- newAddressedBlock(idx, uint64(j), true)
					}
				}(i)
			}
			Eventually(done).Should(BeClosed())
		})

		It("Should supports threshold 5 and observer 5 as parallel committers", func() {
			instance, err := infra.NewBlockCollector(5, 5)
			Expect(err).NotTo(HaveOccurred())

			block := make(chan *infra.AddressedBlock)
			successRateBlock := make(chan *infra.AddressedBlock)
			done := make(chan struct{})

			go infra.CalSuccessRate(5, 10, successRateBlock)
			go instance.Start(context.Background(), block, successRateBlock, done, 10, time.Now(), false)
			for i := 0; i < 5; i++ {
				go func(idx int) {
					for j := 0; j < 10; j++ {
						block <- newAddressedBlock(idx, uint64(j), true)
					}
				}(i)
			}
			Eventually(done).Should(BeClosed())
		})
	})
})

var _ = Describe("CalSuccessRate", func() {

	Context("all txs are correct", func() {
		It("Should supports observer 1 and tx 1", func() {
			rescueStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			blockCh := make(chan *infra.AddressedBlock)
			go func() {
				blockCh <- newAddressedBlock(0, 0, true)
			}()
			infra.CalSuccessRate(1, 1, blockCh)

			w.Close()
			out, _ := ioutil.ReadAll(r)
			os.Stdout = rescueStdout
			Expect(string(out)).Should(ContainSubstring("peer 0 received 1 txs, containing 1 successful txs, and the success rate is 100.00%"))
			Expect(string(out)).Should(ContainSubstring("All peer received 1 txs, containing 1 successful txs, and the success rate is 100.00%"))
		})

		It("Should supports observer 2 and tx 1", func() {
			rescueStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			blockCh := make(chan *infra.AddressedBlock)
			go func() {
				blockCh <- newAddressedBlock(0, 0, true)
				blockCh <- newAddressedBlock(1, 0, true)
			}()
			infra.CalSuccessRate(2, 1, blockCh)

			w.Close()
			out, _ := ioutil.ReadAll(r)
			os.Stdout = rescueStdout
			Expect(string(out)).Should(ContainSubstring("peer 0 received 1 txs, containing 1 successful txs, and the success rate is 100.00%"))
			Expect(string(out)).Should(ContainSubstring("peer 1 received 1 txs, containing 1 successful txs, and the success rate is 100.00%"))
			Expect(string(out)).Should(ContainSubstring("All peer received 2 txs, containing 2 successful txs, and the success rate is 100.00%"))
		})

		It("Should supports observer 2 and tx 2", func() {
			rescueStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			blockCh := make(chan *infra.AddressedBlock)
			go func() {
				blockCh <- newAddressedBlock(0, 0, true)
				blockCh <- newAddressedBlock(1, 0, true)
				blockCh <- newAddressedBlock(0, 1, true)
				blockCh <- newAddressedBlock(1, 1, true)
			}()
			infra.CalSuccessRate(2, 2, blockCh)

			w.Close()
			out, _ := ioutil.ReadAll(r)
			os.Stdout = rescueStdout
			Expect(string(out)).Should(ContainSubstring("peer 0 received 2 txs, containing 2 successful txs, and the success rate is 100.00%"))
			Expect(string(out)).Should(ContainSubstring("peer 1 received 2 txs, containing 2 successful txs, and the success rate is 100.00%"))
			Expect(string(out)).Should(ContainSubstring("All peer received 4 txs, containing 4 successful txs, and the success rate is 100.00%"))
		})
	})

	Context("Not all txs are correct", func() {
		It("Should supports observer 1 and tx 1", func() {
			rescueStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			blockCh := make(chan *infra.AddressedBlock)
			go func() {
				blockCh <- newAddressedBlock(0, 0, false)
			}()
			infra.CalSuccessRate(1, 1, blockCh)

			w.Close()
			out, _ := ioutil.ReadAll(r)
			os.Stdout = rescueStdout
			Expect(string(out)).Should(ContainSubstring("peer 0 received 1 txs, containing 0 successful txs, and the success rate is 0.00%"))
			Expect(string(out)).Should(ContainSubstring("All peer received 1 txs, containing 0 successful txs, and the success rate is 0.00%"))
		})

		It("Should supports observer 2 and tx 1", func() {
			rescueStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			blockCh := make(chan *infra.AddressedBlock)
			go func() {
				blockCh <- newAddressedBlock(0, 0, false)
				blockCh <- newAddressedBlock(1, 0, false)
			}()
			infra.CalSuccessRate(2, 1, blockCh)

			w.Close()
			out, _ := ioutil.ReadAll(r)
			os.Stdout = rescueStdout
			Expect(string(out)).Should(ContainSubstring("peer 0 received 1 txs, containing 0 successful txs, and the success rate is 0.00%"))
			Expect(string(out)).Should(ContainSubstring("peer 1 received 1 txs, containing 0 successful txs, and the success rate is 0.00%"))
			Expect(string(out)).Should(ContainSubstring("All peer received 2 txs, containing 0 successful txs, and the success rate is 0.00%"))
		})

		It("Should supports observer 2 and tx 2", func() {
			rescueStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			blockCh := make(chan *infra.AddressedBlock)
			go func() {
				blockCh <- newAddressedBlock(0, 0, true)
				blockCh <- newAddressedBlock(1, 0, true)
				blockCh <- newAddressedBlock(0, 1, false)
				blockCh <- newAddressedBlock(1, 1, false)
			}()
			infra.CalSuccessRate(2, 2, blockCh)

			w.Close()
			out, _ := ioutil.ReadAll(r)
			os.Stdout = rescueStdout
			Expect(string(out)).Should(ContainSubstring("peer 0 received 2 txs, containing 1 successful txs, and the success rate is 50.00%"))
			Expect(string(out)).Should(ContainSubstring("peer 1 received 2 txs, containing 1 successful txs, and the success rate is 50.00%"))
			Expect(string(out)).Should(ContainSubstring("All peer received 4 txs, containing 2 successful txs, and the success rate is 50.00%"))
		})
	})
})
