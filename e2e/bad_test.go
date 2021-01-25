package e2e

import (
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var (
	mtlsCertFile, mtlsKeyFile *os.File
	tmpDir, tapeBin           string
	tapeSession               *gexec.Session
)

var _ = Describe("Mock test for error input", func() {

	BeforeSuite(func() {
		tmpDir, err := ioutil.TempDir("", "tape-e2e-")
		Expect(err).NotTo(HaveOccurred())

		mtlsCertFile, err = ioutil.TempFile(tmpDir, "mtls-*.crt")
		Expect(err).NotTo(HaveOccurred())

		mtlsKeyFile, err = ioutil.TempFile(tmpDir, "mtls-*.key")
		Expect(err).NotTo(HaveOccurred())

		err = GenerateCertAndKeys(mtlsKeyFile, mtlsCertFile)
		Expect(err).NotTo(HaveOccurred())

		mtlsCertFile.Close()
		mtlsKeyFile.Close()

		tapeBin, err = gexec.Build("../cmd/tape")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if tapeSession != nil && tapeSession.ExitCode() == -1 {
			tapeSession.Kill()
		}
	})

	AfterSuite(func() {
		os.RemoveAll(tmpDir)
		os.Remove(tapeBin)
	})

	Context("E2E with Error Cases", func() {
		When("Command error", func() {
			It("should return unexpected command", func() {
				cmd := exec.Command(tapeBin, "wrongCommand")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("tape: error: unexpected wrongCommand, try --help"))
			})

			It("should return required flag config", func() {
				cmd := exec.Command(tapeBin, "-n", "500")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("tape: error: required flag --config not provided, try --help"))
			})

			It("should return required flag number", func() {
				cmd := exec.Command(tapeBin, "-c", "TestFile")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("tape: error: required flag --number not provided, try --help"))
			})

			It("should return help info", func() {
				cmd := exec.Command(tapeBin, "--help")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("--help  Show context-sensitive help*"))
			})

			It("should return error msg when negative rate", func() {
				config, err := ioutil.TempFile("", "dummy-*.yaml")
				configValue := Values{
					PrivSk:          "N/A",
					SignCert:        "N/A",
					Mtls:            false,
					PeersAddrs:      []string{"N/A"},
					OrdererAddr:     "N/A",
					CommitThreshold: 1,
				}
				GenerateConfigFile(config.Name(), configValue)
				cmd := exec.Command(tapeBin, "-c", config.Name(), "-n", "500", "--rate=-1")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("tape: error: rate must be zero \\(unlimited\\) or positive number\n"))
			})

			It("should return error msg when less than 1 burst", func() {
				cmd := exec.Command(tapeBin, "-c", "config", "-n", "500", "--burst", "0")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("tape: error: burst at least 1\n"))
			})

			It("should return warning msg when rate bigger than burst", func() {
				cmd := exec.Command(tapeBin, "-c", "NoExitFile", "-n", "500", "--rate=10", "--burst", "1")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Out).Should(Say("As rate 10 is bigger than burst 1, real rate is burst\n"))
				Eventually(tapeSession.Err).Should(Say("NoExitFile"))
			})

			It("should return warning msg when rate bigger than default burst", func() {
				cmd := exec.Command(tapeBin, "-c", "NoExitFile", "-n", "500", "--rate", "10000")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Out).Should(Say("As rate 10000 is bigger than burst 1000, real rate is burst\n"))
				Eventually(tapeSession.Err).Should(Say("NoExitFile"))
			})
		})

		When("Config error", func() {
			It("should return file not exist", func() {
				cmd := exec.Command(tapeBin, "-c", "NoExitFile", "-n", "500")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("NoExitFile"))
			})

			It("should return MSP error", func() {
				config, err := ioutil.TempFile("", "dummy-*.yaml")
				configValue := Values{
					PrivSk:          "N/A",
					SignCert:        "N/A",
					Mtls:            false,
					PeersAddrs:      []string{"N/A"},
					OrdererAddr:     "N/A",
					CommitThreshold: 0,
				}
				GenerateConfigFile(config.Name(), configValue)
				cmd := exec.Command(tapeBin, "-c", config.Name(), "-n", "500")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("error loading priv key"))
			})

			It("returns error if commitThreshold is greater than # of committers", func() {
				config, err := ioutil.TempFile("", "no-tls-config-*.yaml")
				configValue := Values{
					PrivSk:          mtlsKeyFile.Name(),
					SignCert:        mtlsCertFile.Name(),
					Mtls:            false,
					PeersAddrs:      []string{"dummy-address"},
					OrdererAddr:     "N/A",
					CommitThreshold: 2,
				}
				GenerateConfigFile(config.Name(), configValue)

				cmd := exec.Command(tapeBin, "-c", config.Name(), "-n", "500")
				tapeSession, err = gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("failed to create block collector"))
			})
		})

		When("Network connection error", func() {
			It("should hit with error", func() {
				config, err := ioutil.TempFile("", "dummy-*.yaml")
				configValue := Values{
					PrivSk:          mtlsKeyFile.Name(),
					SignCert:        mtlsCertFile.Name(),
					Mtls:            false,
					PeersAddrs:      []string{"invalid_addr"},
					OrdererAddr:     "N/A",
					CommitThreshold: 1,
				}
				GenerateConfigFile(config.Name(), configValue)

				cmd := exec.Command(tapeBin, "-c", config.Name(), "-n", "500")
				tapeSession, err = gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("error connecting to invalid_addr"))
			})
		})
	})
})
