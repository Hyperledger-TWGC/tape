package infra

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

var configText = `
{
  "peer_addr": "peer0.org1.example.com:7051",
  "orderer_addr": "orderer.example.com:7050",
  "channel": "mychannel",
  "chaincode": "mycc",
  "args": ["put", "key", "value"],
  "mspid": "Org1MSP",
  "private_key": "/path/to/private.key",
  "sign_cert": "/path/to/sign.cert",
  "tls_ca_certs": ["/path/to/peer/tls/ca/cert","/path/to/orderer/tls/ca/cert"],
  "num_of_conn": 20,
  "client_per_conn": 40
}`

func TestLoadConfig(t *testing.T) {
	f, err := ioutil.TempFile("", "config-*.json")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	_, err = f.WriteString(configText)
	assert.NoError(t, err)
	err = f.Close()
	assert.NoError(t, err)

	c := LoadConfig(f.Name())
	assert.Equal(t, c.OrdererAddr, "orderer.example.com:7050")
	assert.Equal(t, c.Channel, "mychannel")
	assert.Equal(t, c.NumOfConn, 20)
}
