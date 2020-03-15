package infra

import (
	"encoding/json"
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
)

type Config struct {
	PeerAddrs     []string `json:"peer_addrs"`
	OrdererAddr   string   `json:"orderer_addr"`
	Channel       string   `json:"channel"`
	Chaincode     string   `json:"chaincode"`
	Version       string   `json:"version"`
	Args          []string `json:"args"`
	MSPID         string   `json:"mspid"`
	PrivateKey    string   `json:"private_key"`
	SignCert      string   `json:"sign_cert"`
	TLSCACerts    []string `json:"tls_ca_certs"`
	NumOfConn     int      `json:"num_of_conn"`
	ClientPerConn int      `json:"client_per_conn"`
}

func LoadConfig(f string) Config {
	raw, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}

	config := Config{}
	if err = json.Unmarshal(raw, &config); err != nil {
		panic(err)
	}

	return config
}

func (c Config) LoadCrypto() *Crypto {
	conf := CryptoConfig{
		MSPID:      c.MSPID,
		PrivKey:    c.PrivateKey,
		SignCert:   c.SignCert,
		TLSCACerts: c.TLSCACerts,
	}

	priv, err := GetPrivateKey(conf.PrivKey)
	if err != nil {
		panic(err)
	}

	cert, certBytes, err := GetCertificate(conf.SignCert)
	if err != nil {
		panic(err)
	}

	id := &msp.SerializedIdentity{
		Mspid:   conf.MSPID,
		IdBytes: certBytes,
	}

	name, err := proto.Marshal(id)
	if err != nil {
		panic(err)
	}

	certs, err := GetTLSCACerts(conf.TLSCACerts)
	if err != nil {
		panic(err)
	}

	return &Crypto{
		Creator:    name,
		PrivKey:    priv,
		SignCert:   cert,
		TLSCACerts: certs,
	}
}
