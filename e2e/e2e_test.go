package e2e

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"os"
	"os/exec"

	"tape/e2e/mock"

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
					Addr:            "N/A",
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
					Addr:            "N/A",
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
					Addr:            "dummy-address",
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
					Addr:            "invalid_addr",
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

	Context("E2E with correct subcommand", func() {
		When("Version subcommand", func() {
			It("should return version info", func() {
				cmd := exec.Command(tapeBin, "version")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Out).Should(Say("tape:\n Version:.*\n Go version:.*\n OS/Arch:.*\n"))
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
				configValue := Values{
					PrivSk:          mtlsKeyFile.Name(),
					SignCert:        mtlsCertFile.Name(),
					Mtls:            false,
					Addr:            lis.Addr().String(),
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
				configValue := Values{
					PrivSk:          mtlsKeyFile.Name(),
					SignCert:        mtlsCertFile.Name(),
					Mtls:            true,
					MtlsCrt:         mtlsCertFile.Name(),
					MtlsKey:         mtlsKeyFile.Name(),
					Addr:            lis.Addr().String(),
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
				lis, err := net.Listen("tcp", "127.0.0.1:0")
				Expect(err).NotTo(HaveOccurred())

				grpcServer := grpc.NewServer()

				mock := &mock.Server{GrpcServer: grpcServer, Listener: lis}
				go mock.Start()
				defer mock.Stop()

				config, err := ioutil.TempFile("", "Rate-*.yaml")
				configValue := Values{
					PrivSk:          mtlsKeyFile.Name(),
					SignCert:        mtlsCertFile.Name(),
					Mtls:            false,
					Addr:            lis.Addr().String(),
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
				lis, err := net.Listen("tcp", "127.0.0.1:0")
				Expect(err).NotTo(HaveOccurred())

				grpcServer := grpc.NewServer()

				mock := &mock.Server{GrpcServer: grpcServer, Listener: lis}
				go mock.Start()
				defer mock.Stop()

				config, err := ioutil.TempFile("", "burst-*.yaml")
				configValue := Values{
					PrivSk:          mtlsKeyFile.Name(),
					SignCert:        mtlsCertFile.Name(),
					Mtls:            false,
					Addr:            lis.Addr().String(),
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
				lis, err := net.Listen("tcp", "127.0.0.1:0")
				Expect(err).NotTo(HaveOccurred())

				grpcServer := grpc.NewServer()

				mock := &mock.Server{GrpcServer: grpcServer, Listener: lis}
				go mock.Start()
				defer mock.Stop()

				config, err := ioutil.TempFile("", "BothRateAndBurst-*.yaml")
				configValue := Values{
					PrivSk:          mtlsKeyFile.Name(),
					SignCert:        mtlsCertFile.Name(),
					Mtls:            false,
					Addr:            lis.Addr().String(),
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
