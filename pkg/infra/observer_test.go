package infra_test

import (
	"context"
	"io/ioutil"
	"os"
	"tape/e2e"
	"tape/e2e/mock"
	"tape/pkg/infra"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("Observer", func() {
	var (
		tmpDir                    string
		logger                    *log.Logger
		mtlsCertFile, mtlsKeyFile *os.File
	)

	BeforeEach(func() {
		logger = log.New()

		tmpDir, err := ioutil.TempDir("", "tape-")
		Expect(err).NotTo(HaveOccurred())

		mtlsCertFile, err = ioutil.TempFile(tmpDir, "mtls-*.crt")
		Expect(err).NotTo(HaveOccurred())

		mtlsKeyFile, err = ioutil.TempFile(tmpDir, "mtls-*.key")
		Expect(err).NotTo(HaveOccurred())

		err = e2e.GenerateCertAndKeys(mtlsKeyFile, mtlsCertFile)
		Expect(err).NotTo(HaveOccurred())

		mtlsCertFile.Close()
		mtlsKeyFile.Close()
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	It("It should work with mock", func() {
		txC := make(chan struct{}, mock.MockTxSize)
		mpeer, err := mock.NewPeer(txC, nil)
		Expect(err).NotTo(HaveOccurred())
		go mpeer.Start()
		defer mpeer.Stop()
		configFile, err := ioutil.TempFile(tmpDir, "config*.yaml")
		Expect(err).NotTo(HaveOccurred())
		paddrs := make([]string, 0)
		paddrs = append(paddrs, mpeer.Addrs())

		configValue := e2e.Values{
			PrivSk:          mtlsKeyFile.Name(),
			SignCert:        mtlsCertFile.Name(),
			Mtls:            false,
			PeersAddrs:      paddrs,
			OrdererAddr:     "",
			CommitThreshold: 1,
		}
		e2e.GenerateConfigFile(configFile.Name(), configValue)
		config, err := infra.LoadConfig(configFile.Name())
		Expect(err).NotTo(HaveOccurred())
		crypto, err := config.LoadCrypto()
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		observers, err := infra.CreateObservers(ctx, config.Channel, config.Committers, crypto, logger)
		Expect(err).NotTo(HaveOccurred())

		finishCh := make(chan struct{})
		errorCh := make(chan error, 10)
		start := time.Now()
		blockCollector, err := infra.NewBlockCollector(config.CommitThreshold, len(config.Committers))
		Expect(err).NotTo(HaveOccurred())
		blockCh := make(chan *infra.AddressedBlock)
		go blockCollector.Start(ctx, blockCh, finishCh, mock.MockTxSize, time.Now(), true)
		go observers.Start(errorCh, blockCh, start)
		go func() {
			for i := 0; i < mock.MockTxSize; i++ {
				txC <- struct{}{}
			}
		}()
		Eventually(finishCh).Should(BeClosed())
		completed := time.Now()
		Expect(start.Sub(completed)).Should(BeNumerically("<", 0.002), "observer with mock shouldn't take too long.")
	})

	It("It should work as 2 committed of 3 peers", func() {

		TotalPeers := 3
		CommitThreshold := 2
		paddrs := make([]string, 0)
		txCs := make([]chan struct{}, 0)
		var mpeers []*mock.Peer

		for i := 0; i < TotalPeers; i++ {
			txC := make(chan struct{}, mock.MockTxSize)
			mpeer, err := mock.NewPeer(txC, nil)
			Expect(err).NotTo(HaveOccurred())
			go mpeer.Start()
			defer mpeer.Stop()

			paddrs = append(paddrs, mpeer.Addrs())
			mpeers = append(mpeers, mpeer)
			txCs = append(txCs, txC)
		}

		configFile, err := ioutil.TempFile(tmpDir, "config*.yaml")
		Expect(err).NotTo(HaveOccurred())
		configValue := e2e.Values{
			PrivSk:          mtlsKeyFile.Name(),
			SignCert:        mtlsCertFile.Name(),
			Mtls:            false,
			PeersAddrs:      paddrs,
			OrdererAddr:     "",
			CommitThreshold: CommitThreshold,
		}
		e2e.GenerateConfigFile(configFile.Name(), configValue)
		config, err := infra.LoadConfig(configFile.Name())
		Expect(err).NotTo(HaveOccurred())
		crypto, err := config.LoadCrypto()
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		observers, err := infra.CreateObservers(ctx, config.Channel, config.Committers, crypto, logger)
		Expect(err).NotTo(HaveOccurred())

		finishCh := make(chan struct{})
		errorCh := make(chan error, 10)
		start := time.Now()
		blockCollector, err := infra.NewBlockCollector(config.CommitThreshold, len(config.Committers))
		Expect(err).NotTo(HaveOccurred())
		blockCh := make(chan *infra.AddressedBlock)
		go blockCollector.Start(ctx, blockCh, finishCh, mock.MockTxSize, time.Now(), true)
		go observers.Start(errorCh, blockCh, start)
		for i := 0; i < TotalPeers; i++ {
			go func(k int) {
				for j := 0; j < mock.MockTxSize; j++ {
					txCs[k] <- struct{}{}
				}
			}(i)
		}
		for i := 0; i < CommitThreshold; i++ {
			mpeers[i].Pause()
		}
		Consistently(finishCh).ShouldNot(Receive())
		for i := 0; i < CommitThreshold; i++ {
			mpeers[i].Unpause()
		}
		Eventually(finishCh).Should(BeClosed())
	})
})
