package e2e_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Mock test for version", func() {

	Context("E2E with correct subcommand", func() {
		When("Version subcommand", func() {
			It("should return version info", func() {
				cmd := exec.Command(tapeBin, "version")
				tapeSession, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Eventually(tapeSession.Out).Should(Say("tape:\n Version:.*\n Go version:.*\n Git commit:.*\n Built:.*\n OS/Arch:.*\n"))
			})
		})
	})
})
