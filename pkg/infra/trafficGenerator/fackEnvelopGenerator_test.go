package trafficGenerator_test

import (
	"os"
	"time"

	"github.com/hyperledger-twgc/tape/e2e"
	"github.com/hyperledger-twgc/tape/pkg/infra/basic"
	"github.com/hyperledger-twgc/tape/pkg/infra/trafficGenerator"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("FackEnvelopGenerator", func() {

	var (
		configFile *os.File
		tmpDir     string
	)

	BeforeEach(func() {
		tmpDir, err := os.MkdirTemp("", "tape-")
		Expect(err).NotTo(HaveOccurred())

		mtlsCertFile, err := os.CreateTemp(tmpDir, "mtls-*.crt")
		Expect(err).NotTo(HaveOccurred())

		mtlsKeyFile, err := os.CreateTemp(tmpDir, "mtls-*.key")
		Expect(err).NotTo(HaveOccurred())

		err = e2e.GenerateCertAndKeys(mtlsKeyFile, mtlsCertFile)
		Expect(err).NotTo(HaveOccurred())

		PolicyFile, err := os.CreateTemp(tmpDir, "policy")
		Expect(err).NotTo(HaveOccurred())

		err = e2e.GeneratePolicy(PolicyFile)
		Expect(err).NotTo(HaveOccurred())

		mtlsCertFile.Close()
		mtlsKeyFile.Close()
		PolicyFile.Close()

		configFile, err = os.CreateTemp(tmpDir, "config*.yaml")
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

	It("should crete proposal to raw without limit when limit is 0", func() {
		envs := make(chan *basic.TracingEnvelope, 1002)
		defer close(envs)
		errorCh := make(chan error, 1002)
		defer close(errorCh)
		config, err := basic.LoadConfig(configFile.Name())
		Expect(err).NotTo(HaveOccurred())
		crypto, err := config.LoadCrypto()
		Expect(err).NotTo(HaveOccurred())
		t := time.Now()
		fackEnvelopGenerator := &trafficGenerator.FackEnvelopGenerator{
			Num:     1002,
			Burst:   10,
			R:       0,
			Config:  config,
			Crypto:  crypto,
			Envs:    envs,
			ErrorCh: errorCh,
		}
		fackEnvelopGenerator.Start()
		t1 := time.Now()
		Expect(envs).To(HaveLen(1002))
		Expect(t1.Sub(t)).To(BeNumerically("<", 2*time.Second))

	})

	It("should crete proposal to raw with given limit bigger than 0 less than size", func() {
		envs := make(chan *basic.TracingEnvelope, 1002)
		defer close(envs)
		errorCh := make(chan error, 1002)
		defer close(errorCh)
		config, err := basic.LoadConfig(configFile.Name())
		Expect(err).NotTo(HaveOccurred())
		crypto, err := config.LoadCrypto()
		Expect(err).NotTo(HaveOccurred())
		t := time.Now()
		fackEnvelopGenerator := &trafficGenerator.FackEnvelopGenerator{
			Num:     12,
			Burst:   10,
			R:       1,
			Config:  config,
			Crypto:  crypto,
			Envs:    envs,
			ErrorCh: errorCh,
		}
		fackEnvelopGenerator.Start()
		t1 := time.Now()
		Expect(envs).To(HaveLen(12))
		Expect(t1.Sub(t)).To(BeNumerically("<", 2*time.Second))
	})

	It("should crete proposal to raw with given limit bigger than Size", func() {
		envs := make(chan *basic.TracingEnvelope, 1002)
		defer close(envs)
		errorCh := make(chan error, 1002)
		defer close(errorCh)
		config, err := basic.LoadConfig(configFile.Name())
		Expect(err).NotTo(HaveOccurred())
		crypto, err := config.LoadCrypto()
		Expect(err).NotTo(HaveOccurred())
		t := time.Now()
		fackEnvelopGenerator := &trafficGenerator.FackEnvelopGenerator{
			Num:     12,
			Burst:   10,
			R:       0,
			Config:  config,
			Crypto:  crypto,
			Envs:    envs,
			ErrorCh: errorCh,
		}
		fackEnvelopGenerator.Start()
		t1 := time.Now()
		Expect(envs).To(HaveLen(12))
		Expect(t1.Sub(t)).To(BeNumerically("<", 2*time.Second))
	})

})
