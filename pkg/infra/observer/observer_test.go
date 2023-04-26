package observer_test

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/hyperledger-twgc/tape/e2e"
	"github.com/hyperledger-twgc/tape/e2e/mock"
	"github.com/hyperledger-twgc/tape/pkg/infra/basic"
	"github.com/hyperledger-twgc/tape/pkg/infra/observer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("Observer", func() {
	var (
		tmpDir                                string
		logger                                *log.Logger
		PolicyFile, mtlsCertFile, mtlsKeyFile *os.File
	)
	type key string

	const start key = "start"

	BeforeEach(func() {
		logger = log.New()

		tmpDir, err := os.MkdirTemp("", "tape-")
		Expect(err).NotTo(HaveOccurred())

		mtlsCertFile, err = os.CreateTemp(tmpDir, "mtls-*.crt")
		Expect(err).NotTo(HaveOccurred())

		mtlsKeyFile, err = os.CreateTemp(tmpDir, "mtls-*.key")
		Expect(err).NotTo(HaveOccurred())

		PolicyFile, err = os.CreateTemp(tmpDir, "policy")
		Expect(err).NotTo(HaveOccurred())

		err = e2e.GenerateCertAndKeys(mtlsKeyFile, mtlsCertFile)
		Expect(err).NotTo(HaveOccurred())

		err = e2e.GeneratePolicy(PolicyFile)
		Expect(err).NotTo(HaveOccurred())

		PolicyFile.Close()
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
		configFile, err := os.CreateTemp(tmpDir, "config*.yaml")
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
			PolicyFile:      PolicyFile.Name(),
		}
		e2e.GenerateConfigFile(configFile.Name(), configValue)
		config, err := basic.LoadConfig(configFile.Name())
		Expect(err).NotTo(HaveOccurred())
		crypto, err := config.LoadCrypto()
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		ctx = context.WithValue(ctx, start, time.Now())
		defer cancel()
		errorCh := make(chan error, 10)
		blockCh := make(chan *observer.AddressedBlock)

		observers, err := observer.CreateObservers(ctx, crypto, errorCh, blockCh, config, logger)
		Expect(err).NotTo(HaveOccurred())

		finishCh := make(chan struct{})
		var once sync.Once
		blockCollector, err := observer.NewBlockCollector(config.CommitThreshold, len(config.Committers), ctx, blockCh, finishCh, mock.MockTxSize, false, logger, &once, true)
		Expect(err).NotTo(HaveOccurred())
		go blockCollector.Start()
		go observers.Start()
		go func() {
			for i := 0; i < mock.MockTxSize; i++ {
				txC <- struct{}{}
			}
		}()
		Eventually(finishCh).Should(BeClosed())
		completed := time.Now()
		Expect(ctx.Value(start).(time.Time).Sub(completed)).Should(BeNumerically("<", 0.002), "observer with mock shouldn't take too long.")
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

		configFile, err := os.CreateTemp(tmpDir, "config*.yaml")
		Expect(err).NotTo(HaveOccurred())
		configValue := e2e.Values{
			PrivSk:          mtlsKeyFile.Name(),
			SignCert:        mtlsCertFile.Name(),
			Mtls:            false,
			PeersAddrs:      paddrs,
			OrdererAddr:     "",
			CommitThreshold: CommitThreshold,
			PolicyFile:      PolicyFile.Name(),
		}
		e2e.GenerateConfigFile(configFile.Name(), configValue)
		config, err := basic.LoadConfig(configFile.Name())
		Expect(err).NotTo(HaveOccurred())
		crypto, err := config.LoadCrypto()
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		ctx = context.WithValue(ctx, start, time.Now())
		defer cancel()

		blockCh := make(chan *observer.AddressedBlock)
		errorCh := make(chan error, 10)

		observers, err := observer.CreateObservers(ctx, crypto, errorCh, blockCh, config, logger)
		Expect(err).NotTo(HaveOccurred())

		finishCh := make(chan struct{})
		var once sync.Once
		blockCollector, err := observer.NewBlockCollector(config.CommitThreshold, len(config.Committers), ctx, blockCh, finishCh, mock.MockTxSize, true, logger, &once, true)
		Expect(err).NotTo(HaveOccurred())
		go blockCollector.Start()
		go observers.Start()
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
