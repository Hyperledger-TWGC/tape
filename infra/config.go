package infra

import (
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Peers         []Node   `yaml:"peers"`
	Orderer       Node     `yaml:"orderer"`
	Channel       string   `yaml:"channel"`
	Chaincode     string   `yaml:"chaincode"`
	Version       string   `yaml:"version"`
	Args          []string `yaml:"args"`
	MSPID         string   `yaml:"mspid"`
	PrivateKey    string   `yaml:"private_key"`
	SignCert      string   `yaml:"sign_cert"`
	NumOfConn     int      `yaml:"num_of_conn"`
	ClientPerConn int      `yaml:"client_per_conn"`
}

type Node struct {
	Addr      string `yaml:"addr"`
	TLSCACert string `yaml:"tls_ca_cert"`
}

func LoadConfig(f string) Config {
	raw, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}

	config := Config{}
	if err = yaml.Unmarshal(raw, &config); err != nil {
		panic(err)
	}

	return config
}

func (c Config) LoadCrypto() *Crypto {
	var allcerts []string
	for _, p := range c.Peers {
		allcerts = append(allcerts, p.TLSCACert)
	}
	allcerts = append(allcerts, c.Orderer.TLSCACert)

	conf := CryptoConfig{
		MSPID:      c.MSPID,
		PrivKey:    c.PrivateKey,
		SignCert:   c.SignCert,
		TLSCACerts: allcerts,
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
