package e2e

import (
	"io/ioutil"
	"os/exec"

	"tape/e2e/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Mock test for good path", func() {

	Context("E2E with multi mocked Fabric", func() {
		When("endorsement only", func() {
			It("should work properly", func() {
				server, err := mock.NewServer(2, nil)
				server.Start()
				defer server.Stop()

				config, err := ioutil.TempFile("", "endorsement-only-config-*.yaml")
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

				cmd := exec.Command(tapeBin, "endorsementOnly", "-c", config.Name(), "-n", "500")
				tapeSession, err = gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Out).Should(Say("Time.*Tx.*"))
			})
		})
	})
})
