package infra

import (
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
	"gopkg.in/yaml.v2"
)

const (
	ProcessAll   = iota //0
	EndorseOnly         //1
	EnvelopeOnly        //2
)

type Config struct {
	Endorsers     []Node   `yaml:"endorsers"`
	Committer     Node     `yaml:"committer"`
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
	ProcessFlag   int
}

type Node struct {
	Addr          string `yaml:"addr"`
	TLSCACert     string `yaml:"tls_ca_cert"`
	TLSCAKey      string `yaml:"tls_ca_key"`
	TLSCARoot     string `yaml:"tls_ca_root"`
	TLSCACertByte []byte
	TLSCAKeyByte  []byte
	TLSCARootByte []byte
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

	for i := range config.Endorsers {
		config.Endorsers[i].loadConfig()
	}
	config.Committer.loadConfig()
	config.Orderer.loadConfig()
	if len(config.Endorsers) > 0 {
		config.ProcessFlag = EndorseOnly
	}

	if len(config.Orderer.Addr) > 0 {
		config.ProcessFlag = EnvelopeOnly
	}
	if len(config.Endorsers) > 0 && len(config.Orderer.Addr) > 0 {
		config.ProcessFlag = ProcessAll
	}
	return config
}

func (c Config) LoadCrypto() *Crypto {
	var allcerts []string
	for _, p := range c.Endorsers {
		allcerts = append(allcerts, p.TLSCACert)
	}
	allcerts = append(allcerts, c.Orderer.TLSCACert)

	conf := CryptoConfig{
		MSPID:    c.MSPID,
		PrivKey:  c.PrivateKey,
		SignCert: c.SignCert,
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

	return &Crypto{
		Creator:  name,
		PrivKey:  priv,
		SignCert: cert,
	}
}

func GetTLSCACerts(file string) ([]byte, error) {
	if len(file) == 0 {
		return nil, nil
	}

	in, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return in, nil
}

func (n *Node) loadConfig() {
	TLSCACert, err := GetTLSCACerts(n.TLSCACert)
	if err != nil {
		panic(err)
	}
	certPEM, err := GetTLSCACerts(n.TLSCAKey)
	if err != nil {
		panic(err)
	}
	TLSCARoot, err := GetTLSCACerts(n.TLSCARoot)
	if err != nil {
		panic(err)
	}
	n.TLSCACertByte = TLSCACert
	n.TLSCAKeyByte = certPEM
	n.TLSCARootByte = TLSCARoot
}
