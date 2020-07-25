package infra

import (
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
	"gopkg.in/yaml.v2"
)

type Config struct {
	PeerAddrs     []string `yaml:"peer_addrs"`
	OrdererAddr   string   `yaml:"orderer_addr"`
	Channel       string   `yaml:"channel"`
	Chaincode     string   `yaml:"chaincode"`
	Version       string   `yaml:"version"`
	Args          []string `yaml:"args"`
	MSPID         string   `yaml:"mspid"`
	PrivateKey    string   `yaml:"private_key"`
	SignCert      string   `yaml:"sign_cert"`
	TLSCACerts    []string `yaml:"tls_ca_certs"`
	NumOfConn     int      `yaml:"num_of_conn"`
	ClientPerConn int      `yaml:"client_per_conn"`
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
