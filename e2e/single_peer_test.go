package e2e

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os/exec"

	"tape/e2e/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"google.golang.org/grpc/credentials"
)

var _ = Describe("Mock test for good path", func() {

	Context("E2E with mocked Fabric", func() {
		When("TLS is disabled", func() {
			It("should work properly", func() {
				server, err := mock.NewServer(1, nil)
				Expect(err).NotTo(HaveOccurred())
				server.Start()
				defer server.Stop()

				config, err := ioutil.TempFile("", "no-tls-config-*.yaml")
				paddrs, oaddr := server.Addresses()
				configValue := Values{
					PrivSk:          mtlsKeyFile.Name(),
					SignCert:        mtlsCertFile.Name(),
					Mtls:            false,
					PeersAddrs:      paddrs,
					OrdererAddr:     oaddr,
					CommitThreshold: 1,
				}
				GenerateConfigFile(config.Name(), configValue)

				cmd := exec.Command(tapeBin, "-c", config.Name(), "-n", "500")
				tapeSession, err = gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Out).Should(Say("Time.*Block.*Tx.*10.*"))
			})
		})

		When("client authentication is required", func() {
			It("should work properly", func() {
				peerCert, err := tls.LoadX509KeyPair(mtlsCertFile.Name(),
					mtlsKeyFile.Name())
				Expect(err).NotTo(HaveOccurred())

				caCert, err := ioutil.ReadFile(mtlsCertFile.Name())
				Expect(err).NotTo(HaveOccurred())

				caCertPool := x509.NewCertPool()
				caCertPool.AppendCertsFromPEM(caCert)
				credentials := credentials.NewTLS(&tls.Config{
					Certificates: []tls.Certificate{peerCert},
					ClientCAs:    caCertPool,
					ClientAuth:   tls.RequireAndVerifyClientCert,
				})

				server, err := mock.NewServer(1, credentials)
				server.Start()
				defer server.Stop()

				config, err := ioutil.TempFile("", "mtls-config-*.yaml")
				paddrs, oaddr := server.Addresses()

				configValue := Values{
					PrivSk:          mtlsKeyFile.Name(),
					SignCert:        mtlsCertFile.Name(),
					Mtls:            true,
					MtlsCrt:         mtlsCertFile.Name(),
					MtlsKey:         mtlsKeyFile.Name(),
					PeersAddrs:      paddrs,
					OrdererAddr:     oaddr,
					CommitThreshold: 1,
				}

				GenerateConfigFile(config.Name(), configValue)

				cmd := exec.Command(tapeBin, "-c", config.Name(), "-n", "500")
				tapeSession, err = gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Out).Should(Say("Time.*Block.*Tx.*10.*"))
			})
		})

		When("Only rate is specified", func() {
			It("should work properly", func() {
				server, err := mock.NewServer(1, nil)
				server.Start()
				defer server.Stop()

				config, err := ioutil.TempFile("", "Rate-*.yaml")
				paddrs, oaddr := server.Addresses()

				configValue := Values{
					PrivSk:          mtlsKeyFile.Name(),
					SignCert:        mtlsCertFile.Name(),
					Mtls:            false,
					PeersAddrs:      paddrs,
					OrdererAddr:     oaddr,
					CommitThreshold: 1,
				}
				GenerateConfigFile(config.Name(), configValue)

				cmd := exec.Command(tapeBin, "-c", config.Name(), "-n", "500", "--rate", "10")
				tapeSession, err = gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Out).Should(Say("Time.*Block.*Tx.*10.*"))
			})
		})

		When("Only burst is specified", func() {
			It("should work properly", func() {
				server, err := mock.NewServer(1, nil)
				server.Start()
				defer server.Stop()

				config, err := ioutil.TempFile("", "burst-*.yaml")
				paddrs, oaddr := server.Addresses()

				configValue := Values{
					PrivSk:          mtlsKeyFile.Name(),
					SignCert:        mtlsCertFile.Name(),
					Mtls:            false,
					PeersAddrs:      paddrs,
					OrdererAddr:     oaddr,
					CommitThreshold: 1,
				}
				GenerateConfigFile(config.Name(), configValue)

				cmd := exec.Command(tapeBin, "-c", config.Name(), "-n", "500", "--burst", "10")
				tapeSession, err = gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Out).Should(Say("Time.*Block.*Tx.*10.*"))
			})
		})

		//Test with All arguments
		When("Both rate and burst are specificed", func() {
			It("should work properly", func() {
				server, err := mock.NewServer(1, nil)
				server.Start()
				defer server.Stop()

				config, err := ioutil.TempFile("", "BothRateAndBurst-*.yaml")
				paddrs, oaddr := server.Addresses()

				configValue := Values{
					PrivSk:          mtlsKeyFile.Name(),
					SignCert:        mtlsCertFile.Name(),
					Mtls:            false,
					PeersAddrs:      paddrs,
					OrdererAddr:     oaddr,
					CommitThreshold: 1,
				}
				GenerateConfigFile(config.Name(), configValue)

				cmd := exec.Command(tapeBin, "-c", config.Name(), "-n", "500", "--burst", "100", "--rate", "10")
				tapeSession, err = gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Out).Should(Say("Time.*Block.*Tx.*10.*"))
			})
		})
	})
})
