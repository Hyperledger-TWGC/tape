package e2e_test

import (
	"os"
	"os/exec"

	"github.com/hyperledger-twgc/tape/e2e"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Mock test for error input", func() {

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

			PIt("should return required flag number", func() {
				cmd := exec.Command(tapeBin, "-c", "TestFile")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("tape: error: required flag --number not provided, try --help"))
			})

			It("should return help info", func() {
				cmd := exec.Command(tapeBin, "--help")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("usage: tape .*flags.* .*command.* .*args.*"))
			})

			It("should return error msg when negative rate", func() {
				config, err := os.CreateTemp("", "dummy-*.yaml")
				Expect(err).NotTo(HaveOccurred())
				configValue := e2e.Values{
					PrivSk:          "N/A",
					SignCert:        "N/A",
					Mtls:            false,
					PeersAddrs:      []string{"N/A"},
					OrdererAddr:     "N/A",
					CommitThreshold: 1,
					PolicyFile:      "N/A",
				}
				e2e.GenerateConfigFile(config.Name(), configValue)
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

			It("should return error msg when less than 1 signerNumber", func() {
				cmd := exec.Command(tapeBin, "-c", "config", "-n", "500", "--signers", "0")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("tape: error: signerNumber at least 1\n"))
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
				config, err := os.CreateTemp("", "dummy-*.yaml")
				Expect(err).NotTo(HaveOccurred())
				configValue := e2e.Values{
					PrivSk:          "N/A",
					SignCert:        "N/A",
					Mtls:            false,
					PeersAddrs:      []string{"N/A"},
					OrdererAddr:     "N/A",
					CommitThreshold: 0,
					PolicyFile:      PolicyFile.Name(),
				}
				e2e.GenerateConfigFile(config.Name(), configValue)
				cmd := exec.Command(tapeBin, "-c", config.Name(), "-n", "500")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("error loading priv key"))
			})

			It("returns error if commitThreshold is greater than # of committers", func() {
				config, err := os.CreateTemp("", "no-tls-config-*.yaml")
				Expect(err).NotTo(HaveOccurred())
				configValue := e2e.Values{
					PrivSk:          mtlsKeyFile.Name(),
					SignCert:        mtlsCertFile.Name(),
					Mtls:            false,
					PeersAddrs:      []string{"dummy-address"},
					OrdererAddr:     "N/A",
					CommitThreshold: 2,
					PolicyFile:      PolicyFile.Name(),
				}
				e2e.GenerateConfigFile(config.Name(), configValue)

				cmd := exec.Command(tapeBin, "-c", config.Name(), "-n", "500")
				tapeSession, err = gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("failed to create block collector"))
			})
		})

		When("Network connection error", func() {
			It("should hit with error", func() {
				config, err := os.CreateTemp("", "dummy-*.yaml")
				Expect(err).NotTo(HaveOccurred())
				configValue := e2e.Values{
					PrivSk:          mtlsKeyFile.Name(),
					SignCert:        mtlsCertFile.Name(),
					Mtls:            false,
					PeersAddrs:      []string{"invalid_addr"},
					OrdererAddr:     "N/A",
					CommitThreshold: 1,
					PolicyFile:      PolicyFile.Name(),
				}
				e2e.GenerateConfigFile(config.Name(), configValue)

				cmd := exec.Command(tapeBin, "-c", config.Name(), "-n", "500")
				tapeSession, err = gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("error connecting to invalid_addr"))
			})
		})
	})
})
