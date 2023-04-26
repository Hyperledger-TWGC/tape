package e2e_test

import (
	"os"
	"os/exec"

	"github.com/hyperledger-twgc/tape/e2e"
	"github.com/hyperledger-twgc/tape/e2e/mock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Mock test for good path", func() {

	Context("E2E with multi mocked Fabric", func() {
		When("envelope only", func() {
			PIt("should work properly", func() {
				server, err := mock.NewServer(2, nil)
				Expect(err).NotTo(HaveOccurred())
				server.Start()
				defer server.Stop()

				config, err := os.CreateTemp("", "envelop-only-config-*.yaml")
				Expect(err).NotTo(HaveOccurred())
				paddrs, oaddr := server.Addresses()
				configValue := e2e.Values{
					PrivSk:          mtlsKeyFile.Name(),
					SignCert:        mtlsCertFile.Name(),
					Mtls:            false,
					PeersAddrs:      paddrs,
					OrdererAddr:     oaddr,
					CommitThreshold: 1,
				}
				e2e.GenerateConfigFile(config.Name(), configValue)

				cmd := exec.Command(tapeBin, "commitOnly", "-c", config.Name(), "-n", "500")
				tapeSession, err = gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Err).Should(Say("Time.*Block.*Tx.*"))

				Eventually(tapeSession.Out).Should(Say("Time.*Block.*Tx.*"))
			})
		})
	})
})
