package infra_test

import (
	"io/ioutil"
	"os"
	"text/template"

	"github.com/guoger/stupid/pkg/infra"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type values struct {
	TlsFile   string
	Endorsers bool
	Committer bool
	Orderer   bool
}

func generateConfigFile(FileName string, values values) {
	var Text = `# Definition of nodes
node: &node
  addr: node:port
  tls_ca_cert: {{.TlsFile}}

{{ if .Endorsers }}
endorsers:
  - *node
  - *node
{{ end }}
{{ if .Committer }}
committer: *node
{{ end }}
{{ if .Orderer }}
orderer: *node
{{ end }}

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

	Context("config", func() {
		When("Endorsers,Committer,Orderer given", func() {
			It("successful load", func() {
				tlsFile, err := ioutil.TempFile("", "dummy-*.pem")
				Expect(err).NotTo(HaveOccurred())
				defer os.Remove(tlsFile.Name())

				_, err = tlsFile.Write([]byte("a"))
				Expect(err).NotTo(HaveOccurred())

				f, _ := ioutil.TempFile("", "config-*.yaml")
				defer os.Remove(f.Name())

				generateConfigFile(f.Name(), values{
					TlsFile:   tlsFile.Name(),
					Endorsers: true,
					Committer: true,
					Orderer:   true,
				})

				c := infra.LoadConfig(f.Name())

				Expect(c).To(Equal(infra.Config{
					Endorsers: []infra.Node{
						{Addr: "node:port", TLSCACert: tlsFile.Name(), TLSCACertByte: []byte("a")},
						{Addr: "node:port", TLSCACert: tlsFile.Name(), TLSCACertByte: []byte("a")},
					},
					Committer:     infra.Node{Addr: "node:port", TLSCACert: tlsFile.Name(), TLSCACertByte: []byte("a")},
					Orderer:       infra.Node{Addr: "node:port", TLSCACert: tlsFile.Name(), TLSCACertByte: []byte("a")},
					Channel:       "mychannel",
					Chaincode:     "mycc",
					Version:       "",
					Args:          []string{"invoke", "a", "b", "1"},
					MSPID:         "Org1MSP",
					PrivateKey:    "/path/to/private.key",
					SignCert:      "/path/to/sign.cert",
					NumOfConn:     20,
					ClientPerConn: 40,
					ProcessFlag:   infra.ProcessAll,
				}))
			})
		})
		When("Endorsers given", func() {
			It("successful load", func() {
				tlsFile, err := ioutil.TempFile("", "dummy-*.pem")
				Expect(err).NotTo(HaveOccurred())
				defer os.Remove(tlsFile.Name())

				_, err = tlsFile.Write([]byte("a"))
				Expect(err).NotTo(HaveOccurred())

				f, _ := ioutil.TempFile("", "config-*.yaml")
				defer os.Remove(f.Name())

				generateConfigFile(f.Name(), values{
					TlsFile:   tlsFile.Name(),
					Endorsers: true,
				})

				c := infra.LoadConfig(f.Name())

				Expect(c).To(Equal(infra.Config{
					Endorsers: []infra.Node{
						{Addr: "node:port", TLSCACert: tlsFile.Name(), TLSCACertByte: []byte("a")},
						{Addr: "node:port", TLSCACert: tlsFile.Name(), TLSCACertByte: []byte("a")},
					},
					Committer:     infra.Node{},
					Orderer:       infra.Node{},
					Channel:       "mychannel",
					Chaincode:     "mycc",
					Version:       "",
					Args:          []string{"invoke", "a", "b", "1"},
					MSPID:         "Org1MSP",
					PrivateKey:    "/path/to/private.key",
					SignCert:      "/path/to/sign.cert",
					NumOfConn:     20,
					ClientPerConn: 40,
					ProcessFlag:   infra.EndorseOnly,
				}))
			})
		})
		When("Orderer given", func() {
			It("successful load", func() {
				tlsFile, err := ioutil.TempFile("", "dummy-*.pem")
				Expect(err).NotTo(HaveOccurred())
				defer os.Remove(tlsFile.Name())

				_, err = tlsFile.Write([]byte("a"))
				Expect(err).NotTo(HaveOccurred())

				f, _ := ioutil.TempFile("", "config-*.yaml")
				defer os.Remove(f.Name())

				generateConfigFile(f.Name(), values{
					TlsFile: tlsFile.Name(),
					Orderer: true,
				})
				c := infra.LoadConfig(f.Name())

				Expect(c).To(Equal(infra.Config{
					Endorsers:     nil,
					Committer:     infra.Node{},
					Orderer:       infra.Node{Addr: "node:port", TLSCACert: tlsFile.Name(), TLSCACertByte: []byte("a")},
					Channel:       "mychannel",
					Chaincode:     "mycc",
					Version:       "",
					Args:          []string{"invoke", "a", "b", "1"},
					MSPID:         "Org1MSP",
					PrivateKey:    "/path/to/private.key",
					SignCert:      "/path/to/sign.cert",
					NumOfConn:     20,
					ClientPerConn: 40,
					ProcessFlag:   infra.EnvelopeOnly,
				}))
			})
		})
	})
})
