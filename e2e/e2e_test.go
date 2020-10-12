package e2e

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"os"
	"os/exec"

	"github.com/guoger/tape/e2e/mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var _ = Describe("Mock test", func() {
	var (
		mtlsCertFile, mtlsKeyFile *os.File
		tmpDir, tapeBin           string
		tapeSession               *gexec.Session
	)

	BeforeSuite(func() {
		tmpDir, err := ioutil.TempDir("", "tape-e2e-")
		Expect(err).NotTo(HaveOccurred())

		mtlsCertFile, err = ioutil.TempFile(tmpDir, "mtls-*.crt")
		Expect(err).NotTo(HaveOccurred())

		mtlsKeyFile, err = ioutil.TempFile(tmpDir, "mtls-*.key")
		Expect(err).NotTo(HaveOccurred())

		err = generateCertAndKeys(mtlsKeyFile, mtlsCertFile)
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
		When("Config error", func() {
			It("should hit usage", func() {
				cmd := exec.Command(tapeBin)
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("error input parameters for tape: tape config.yaml 500"))
			})
			It("should hit if not number", func() {
				cmd := exec.Command(tapeBin, "NoExitFile", "abc")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("error input parameters for tape: tape config.yaml 500"))
			})
			It("should return file not exist", func() {
				cmd := exec.Command(tapeBin, "NoExitFile", "500")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("NoExitFile"))
			})

			It("should return MSP error", func() {
				config, err := ioutil.TempFile("", "no-tls-config-*.yaml")
				configValue := values{
					PrivSk:   "N/A",
					SignCert: "N/A",
					Mtls:     false,
					Addr:     "N/A",
				}
				generateConfigFile(config.Name(), configValue)
				cmd := exec.Command(tapeBin, config.Name(), "500")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("error loading priv key"))
			})
		})

		When("Network connection error", func() {
			It("should hit with error", func() {
				config, err := ioutil.TempFile("", "no-tls-config-*.yaml")
				configValue := values{
					PrivSk:   mtlsKeyFile.Name(),
					SignCert: mtlsCertFile.Name(),
					Mtls:     false,
					Addr:     "invalid_addr",
				}
				generateConfigFile(config.Name(), configValue)

				cmd := exec.Command(tapeBin, config.Name(), "500")
				tapeSession, err = gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("error connecting to invalid_addr"))
			})
		})
	})

	Context("E2E with mocked Fabric", func() {
		When("TLS is disabled", func() {
			It("should work properly", func() {
				lis, err := net.Listen("tcp", "127.0.0.1:0")
				Expect(err).NotTo(HaveOccurred())

				grpcServer := grpc.NewServer()

				mock := &mock.Server{GrpcServer: grpcServer, Listener: lis}
				go mock.Start()
				defer mock.Stop()

				config, err := ioutil.TempFile("", "no-tls-config-*.yaml")
				configValue := values{
					PrivSk:   mtlsKeyFile.Name(),
					SignCert: mtlsCertFile.Name(),
					Mtls:     false,
					Addr:     lis.Addr().String(),
				}
				generateConfigFile(config.Name(), configValue)

				cmd := exec.Command(tapeBin, config.Name(), "500")
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
				ta := credentials.NewTLS(&tls.Config{
					Certificates: []tls.Certificate{peerCert},
					ClientCAs:    caCertPool,
					ClientAuth:   tls.RequireAndVerifyClientCert,
				})
				grpcServer := grpc.NewServer(grpc.Creds(ta))

				lis, err := net.Listen("tcp", "127.0.0.1:0")
				Expect(err).NotTo(HaveOccurred())

				mock := &mock.Server{GrpcServer: grpcServer, Listener: lis}
				go mock.Start()
				defer mock.Stop()

				config, err := ioutil.TempFile("", "mtls-config-*.yaml")
				configValue := values{
					PrivSk:   mtlsKeyFile.Name(),
					SignCert: mtlsCertFile.Name(),
					Mtls:     true,
					MtlsCrt:  mtlsCertFile.Name(),
					MtlsKey:  mtlsKeyFile.Name(),
					Addr:     lis.Addr().String(),
				}

				generateConfigFile(config.Name(), configValue)

				cmd := exec.Command(tapeBin, config.Name(), "500")
				tapeSession, err = gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Out).Should(Say("Time.*Block.*Tx.*10.*"))
			})
		})
	})
})
