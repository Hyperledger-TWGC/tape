package trafficGenerator_test

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/Hyperledger-TWGC/tape/e2e"
	"github.com/Hyperledger-TWGC/tape/pkg/infra/basic"
	"github.com/Hyperledger-TWGC/tape/pkg/infra/trafficGenerator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("Initiator", func() {

	var (
		configFile *os.File
		tmpDir     string
		logger     = log.New()
	)

	BeforeEach(func() {
		tmpDir, err := ioutil.TempDir("", "tape-")
		Expect(err).NotTo(HaveOccurred())

		mtlsCertFile, err := ioutil.TempFile(tmpDir, "mtls-*.crt")
		Expect(err).NotTo(HaveOccurred())

		mtlsKeyFile, err := ioutil.TempFile(tmpDir, "mtls-*.key")
		Expect(err).NotTo(HaveOccurred())

		PolicyFile, err := ioutil.TempFile(tmpDir, "policy")
		Expect(err).NotTo(HaveOccurred())

		err = e2e.GeneratePolicy(PolicyFile)
		Expect(err).NotTo(HaveOccurred())

		err = e2e.GenerateCertAndKeys(mtlsKeyFile, mtlsCertFile)
		Expect(err).NotTo(HaveOccurred())

		mtlsCertFile.Close()
		mtlsKeyFile.Close()
		PolicyFile.Close()

		configFile, err = ioutil.TempFile(tmpDir, "config*.yaml")
		Expect(err).NotTo(HaveOccurred())
		configValue := e2e.Values{
			PrivSk:          mtlsKeyFile.Name(),
			SignCert:        mtlsCertFile.Name(),
			Mtls:            false,
			PeersAddrs:      []string{"dummy"},
			OrdererAddr:     "dummy",
			CommitThreshold: 1,
			PolicyFile:      PolicyFile.Name(),
		}
		e2e.GenerateConfigFile(configFile.Name(), configValue)
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	PIt("should crete proposal to raw without limit when number is 0", func() {
		raw := make(chan *basic.TracingProposal, 1002)
		//defer close(raw)
		errorCh := make(chan error, 1002)
		defer close(errorCh)
		config, err := basic.LoadConfig(configFile.Name())
		Expect(err).NotTo(HaveOccurred())
		crypto, err := config.LoadCrypto()
		Expect(err).NotTo(HaveOccurred())
		Initiator := &trafficGenerator.Initiator{0, 10, 0, config, crypto, logger, raw, errorCh}
		go Initiator.Start()
		for i := 0; i < 1002; i++ {
			_, flag := <-raw
			Expect(flag).To(BeFalse())
		}
		close(raw)
	})

	It("should crete proposal to raw without limit when limit is 0", func() {
		raw := make(chan *basic.TracingProposal, 1002)
		defer close(raw)
		errorCh := make(chan error, 1002)
		defer close(errorCh)
		config, err := basic.LoadConfig(configFile.Name())
		Expect(err).NotTo(HaveOccurred())
		crypto, err := config.LoadCrypto()
		Expect(err).NotTo(HaveOccurred())
		t := time.Now()
		Initiator := &trafficGenerator.Initiator{1002, 10, 0, config, crypto, logger, raw, errorCh}
		Initiator.Start()
		t1 := time.Now()
		Expect(raw).To(HaveLen(1002))
		Expect(t1.Sub(t)).To(BeNumerically("<", 2*time.Second))
	})

	It("should crete proposal to raw with given limit bigger than 0 less than size", func() {
		raw := make(chan *basic.TracingProposal, 1002)
		defer close(raw)
		errorCh := make(chan error, 1002)
		defer close(errorCh)
		config, err := basic.LoadConfig(configFile.Name())
		Expect(err).NotTo(HaveOccurred())
		crypto, err := config.LoadCrypto()
		Expect(err).NotTo(HaveOccurred())
		t := time.Now()
		Initiator := &trafficGenerator.Initiator{12, 10, 1, config, crypto, logger, raw, errorCh}
		Initiator.Start()
		t1 := time.Now()
		Expect(raw).To(HaveLen(12))
		Expect(t1.Sub(t)).To(BeNumerically(">", 2*time.Second))
	})

	It("should crete proposal to raw with given limit bigger than Size", func() {
		raw := make(chan *basic.TracingProposal, 1002)
		defer close(raw)
		errorCh := make(chan error, 1002)
		defer close(errorCh)
		config, err := basic.LoadConfig(configFile.Name())
		Expect(err).NotTo(HaveOccurred())
		crypto, err := config.LoadCrypto()
		Expect(err).NotTo(HaveOccurred())
		t := time.Now()
		Initiator := &trafficGenerator.Initiator{12, 10, 0, config, crypto, logger, raw, errorCh}
		Initiator.Start()
		t1 := time.Now()
		Expect(raw).To(HaveLen(12))
		Expect(t1.Sub(t)).To(BeNumerically("<", 2*time.Second))
	})
})
