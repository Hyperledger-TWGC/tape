package e2e_test

import (
	"os"
	"testing"

	"github.com/hyperledger-twgc/tape/e2e"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var (
	mtlsCertFile, mtlsKeyFile, PolicyFile *os.File
	tmpDir, tapeBin                       string
	tapeSession                           *gexec.Session
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var _ = BeforeSuite(func() {
	tmpDir, err := os.MkdirTemp("", "tape-e2e-")
	Expect(err).NotTo(HaveOccurred())

	mtlsCertFile, err = os.CreateTemp(tmpDir, "mtls-*.crt")
	Expect(err).NotTo(HaveOccurred())

	mtlsKeyFile, err = os.CreateTemp(tmpDir, "mtls-*.key")
	Expect(err).NotTo(HaveOccurred())

	err = e2e.GenerateCertAndKeys(mtlsKeyFile, mtlsCertFile)
	Expect(err).NotTo(HaveOccurred())

	PolicyFile, err = os.CreateTemp(tmpDir, "policy")
	Expect(err).NotTo(HaveOccurred())

	err = e2e.GeneratePolicy(PolicyFile)
	Expect(err).NotTo(HaveOccurred())

	mtlsCertFile.Close()
	mtlsKeyFile.Close()
	PolicyFile.Close()

	tapeBin, err = gexec.Build("../cmd/tape")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterEach(func() {
	if tapeSession != nil && tapeSession.ExitCode() == -1 {
		tapeSession.Kill()
	}
})

var _ = AfterSuite(func() {
	os.RemoveAll(tmpDir)
	os.Remove(tapeBin)
})
