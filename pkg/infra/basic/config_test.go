package basic_test

import (
	"os"
	"text/template"

	"github.com/hyperledger-twgc/tape/pkg/infra/basic"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type files struct {
	TlsFile    string
	PolicyFile string
}

func generateConfigFile(FileName string, values interface{}) {
	var Text = `# Definition of nodes
org1peer0: &org1peer0
  addr: localhost:7051
  ssl_target_name_override: peer0.org1.example.com
  tls_ca_cert: {{.TlsFile}}
  org: org1
org2peer0: &org2peer0
  addr: peer0.org2.example.com:7051
  tls_ca_cert: {{.TlsFile}}
  org: org2
org0orderer0: &org0orderer0
  addr: orderer.example.com:7050
  tls_ca_cert: {{.TlsFile}}
  org: org0

endorsers:
  - *org1peer0
  - *org2peer0
committers: 
  - *org2peer0
commitThreshold: 1
orderer: *org0orderer0
policyFile: {{.PolicyFile}}

channel: mychannel
chaincode: mycc
args:
  - invoke
  - a
  - b
  - 1
mspid: Org1MSP
private_key: /path/to/private.key
sign_cert: /path/to/sign.cert
num_of_conn: 20
client_per_conn: 40
`
	tmpl, err := template.New("test").Parse(Text)
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile(FileName, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = tmpl.Execute(file, values)
	if err != nil {
		panic(err)
	}
}

var _ = Describe("Config", func() {

	Context("good", func() {
		It("successful loads", func() {
			tlsFile, err := os.CreateTemp("", "dummy-*.pem")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(tlsFile.Name())

			_, err = tlsFile.Write([]byte("a"))
			Expect(err).NotTo(HaveOccurred())

			policyFile, err := os.CreateTemp("", "dummy-*.pem")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(policyFile.Name())

			_, err = policyFile.Write([]byte("b"))
			Expect(err).NotTo(HaveOccurred())

			f, _ := os.CreateTemp("", "config-*.yaml")
			defer os.Remove(f.Name())

			file := files{TlsFile: tlsFile.Name(),
				PolicyFile: policyFile.Name()}

			generateConfigFile(f.Name(), file)

			c, err := basic.LoadConfig(f.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(c).To(Equal(basic.Config{
				Endorsers: []basic.Node{
					{Addr: "localhost:7051", SslTargetNameOverride: "peer0.org1.example.com", TLSCACert: tlsFile.Name(), TLSCACertByte: []byte("a"), Org: "org1"},
					{Addr: "peer0.org2.example.com:7051", TLSCACert: tlsFile.Name(), TLSCACertByte: []byte("a"), Org: "org2"},
				},
				Committers:      []basic.Node{{Addr: "peer0.org2.example.com:7051", TLSCACert: tlsFile.Name(), TLSCACertByte: []byte("a"), Org: "org2"}},
				CommitThreshold: 1,
				Orderer:         basic.Node{Addr: "orderer.example.com:7050", TLSCACert: tlsFile.Name(), TLSCACertByte: []byte("a"), Org: "org0"},
				PolicyFile:      policyFile.Name(),
				Rule:            "b",
				Channel:         "mychannel",
				Chaincode:       "mycc",
				Version:         "",
				Args:            []string{"invoke", "a", "b", "1"},
				MSPID:           "Org1MSP",
				PrivateKey:      "/path/to/private.key",
				SignCert:        "/path/to/sign.cert",
				NumOfConn:       20,
				ClientPerConn:   40,
			}))
			_, err = c.LoadCrypto()
			Expect(err).Should(MatchError(ContainSubstring("error loading priv key")))
		})
	})

	Context("bad", func() {
		It("fails to load missing config file", func() {
			_, err := basic.LoadConfig("invalid_file")
			Expect(err).Should(MatchError(ContainSubstring("invalid_file")))
		})

		It("fails to load invalid config file", func() {

			f, _ := os.CreateTemp("", "config-*.yaml")
			defer os.Remove(f.Name())

			file := files{TlsFile: "invalid_file",
				PolicyFile: "invalid_file"}

			generateConfigFile(f.Name(), file)

			_, err := basic.LoadConfig(f.Name())
			Expect(err).Should(MatchError(ContainSubstring("invalid_file")))
		})
	})
})
