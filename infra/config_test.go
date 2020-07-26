package infra_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/guoger/stupid/infra"
)

var _ = Describe("Config", func() {

	Context("config", func() {
		It("successful load", func() {
			var configText = `
org1peer0: &org1peer0
  addr: peer0.org1.example.com:7051
  tls_ca_cert: /path/to/org1peer0/tls/ca/cert
org2peer0: &org2peer0
  addr: peer0.org2.example.com:7051
  tls_ca_cert: /path/to/org2peer0/tls/ca/cert
org0orderer0: &org0orderer0
  addr: orderer.example.com:7050
  tls_ca_cert: /path/to/orderer/tls/ca/cert

peers:
  - *org1peer0
  - *org2peer0

orderer: *org0orderer0

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
client_per_conn: 40`

			f, _ := ioutil.TempFile("", "config-*.yaml")
			defer os.Remove(f.Name())
			f.WriteString(configText)
			f.Close()

			c := infra.LoadConfig(f.Name())

			Expect(c).To(Equal(infra.Config{
				Peers: []infra.Node{
					{Addr: "peer0.org1.example.com:7051", TLSCACert: "/path/to/org1peer0/tls/ca/cert"},
					{Addr: "peer0.org2.example.com:7051", TLSCACert: "/path/to/org2peer0/tls/ca/cert"},
				},
				Orderer:       infra.Node{Addr: "orderer.example.com:7050", TLSCACert: "/path/to/orderer/tls/ca/cert"},
				Channel:       "mychannel",
				Chaincode:     "mycc",
				Version:       "",
				Args:          []string{"invoke", "a", "b", "1"},
				MSPID:         "Org1MSP",
				PrivateKey:    "/path/to/private.key",
				SignCert:      "/path/to/sign.cert",
				NumOfConn:     20,
				ClientPerConn: 40,
			}))
		})
	})
})
