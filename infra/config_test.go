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
			{
			  "peer_addr": "peer0.org1.example.com:7051",
			  "orderer_addr": "orderer.example.com:7050",
			  "channel": "mychannel",
			  "chaincode": "mycc",
			  "args": ["invoke", "a", "b", "1"],
			  "mspid": "Org1MSP",
			  "private_key": "/path/to/private.key",
			  "sign_cert": "/path/to/sign.cert",
			  "tls_ca_certs": ["/path/to/peer/tls/ca/cert","/path/to/orderer/tls/ca/cert"],
			  "num_of_conn": 20,
			  "client_per_conn": 40
			}`
			f, _ := ioutil.TempFile("", "config-*.json")
			defer os.Remove(f.Name())
			f.WriteString(configText)
			f.Close()

			c := infra.LoadConfig(f.Name())
			Expect(c.OrdererAddr).To(Equal("orderer.example.com:7050"))
			Expect(c.Channel).To(Equal("mychannel"))
			Expect(c.NumOfConn).To(Equal(20))
		})
	})
})
