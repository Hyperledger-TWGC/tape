package basic

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"sync"

	"github.com/opentracing/opentracing-go"

	"github.com/hyperledger-twgc/tape/internal/fabric/bccsp/utils"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type TracingProposal struct {
	*peer.Proposal
	TxId string
	Span opentracing.Span
}

type Elements struct {
	TxId            string
	Span            opentracing.Span
	EndorsementSpan opentracing.Span
	SignedProp      *peer.SignedProposal
	Responses       []*peer.ProposalResponse
	Orgs            []string
	Processed       bool
	Lock            sync.Mutex
}

type TracingEnvelope struct {
	Env  *common.Envelope
	TxId string
	Span opentracing.Span
}

type Config struct {
	Endorsers       []Node `yaml:"endorsers"`
	Committers      []Node `yaml:"committers"`
	CommitThreshold int    `yaml:"commitThreshold"`
	Orderer         Node   `yaml:"orderer"`
	PolicyFile      string `yaml:"policyFile"`
	Rule            string
	Channel         string   `yaml:"channel"`
	Chaincode       string   `yaml:"chaincode"`
	Version         string   `yaml:"version"`
	Args            []string `yaml:"args"`
	MSPID           string   `yaml:"mspid"`
	PrivateKey      string   `yaml:"private_key"`
	SignCert        string   `yaml:"sign_cert"`
	NumOfConn       int      `yaml:"num_of_conn"`
	ClientPerConn   int      `yaml:"client_per_conn"`
}

type Node struct {
	Addr                  string `yaml:"addr"`
	SslTargetNameOverride string `yaml:"ssl_target_name_override"`
	TLSCACert             string `yaml:"tls_ca_cert"`
	Org                   string `yaml:"org"`
	TLSCAKey              string `yaml:"tls_ca_key"`
	TLSCARoot             string `yaml:"tls_ca_root"`
	TLSCACertByte         []byte
	TLSCAKeyByte          []byte
	TLSCARootByte         []byte
}

func LoadConfig(f string) (Config, error) {
	config := Config{}
	raw, err := os.ReadFile(f)
	if err != nil {
		return config, errors.Wrapf(err, "error loading %s", f)
	}
	err = yaml.Unmarshal(raw, &config)
	if err != nil {
		return config, errors.Wrapf(err, "error unmarshal %s", f)
	}

	if len(config.PolicyFile) == 0 && config.PolicyFile == "" {
		return config, errors.New("empty endorsement policy")
	}

	// config.Rule read from PolicyFile
	in, err := os.ReadFile(config.PolicyFile)
	if err != nil {
		return config, err
	}
	config.Rule = string(in)

	for i := range config.Endorsers {
		err = config.Endorsers[i].LoadConfig()
		if err != nil {
			return config, err
		}
	}
	for i := range config.Committers {
		err = config.Committers[i].LoadConfig()
		if err != nil {
			return config, err
		}
	}
	err = config.Orderer.LoadConfig()
	if err != nil {
		return config, err
	}
	return config, nil
}

func (c Config) LoadCrypto() (*CryptoImpl, error) {
	conf := CryptoConfig{
		MSPID:    c.MSPID,
		PrivKey:  c.PrivateKey,
		SignCert: c.SignCert,
	}

	priv, err := GetPrivateKey(conf.PrivKey)
	if err != nil {
		return nil, errors.Wrapf(err, "error loading priv key")
	}

	cert, certBytes, err := GetCertificate(conf.SignCert)
	if err != nil {
		return nil, errors.Wrapf(err, "error loading certificate")
	}

	id := &msp.SerializedIdentity{
		Mspid:   conf.MSPID,
		IdBytes: certBytes,
	}

	name, err := proto.Marshal(id)
	if err != nil {
		return nil, errors.Wrapf(err, "error get msp id")
	}

	return &CryptoImpl{
		Creator:  name,
		PrivKey:  priv,
		SignCert: cert,
	}, nil
}

func GetTLSCACerts(file string) ([]byte, error) {
	if len(file) == 0 {
		return nil, nil
	}

	in, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "error loading %s", file)
	}
	return in, nil
}

func (n *Node) LoadConfig() error {
	TLSCACert, err := GetTLSCACerts(n.TLSCACert)
	if err != nil {
		return errors.Wrapf(err, "fail to load TLS CA Cert %s", n.TLSCACert)
	}
	certPEM, err := GetTLSCACerts(n.TLSCAKey)
	if err != nil {
		return errors.Wrapf(err, "fail to load TLS CA Key %s", n.TLSCAKey)
	}
	TLSCARoot, err := GetTLSCACerts(n.TLSCARoot)
	if err != nil {
		return errors.Wrapf(err, "fail to load TLS CA Root %s", n.TLSCARoot)
	}
	n.TLSCACertByte = TLSCACert
	n.TLSCAKeyByte = certPEM
	n.TLSCARootByte = TLSCARoot
	return nil
}

func GetPrivateKey(f string) (*ecdsa.PrivateKey, error) {
	in, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}

	k, err := utils.PEMtoPrivateKey(in, []byte{})
	if err != nil {
		return nil, err
	}

	key, ok := k.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.Errorf("expecting ecdsa key")
	}

	return key, nil
}

func GetCertificate(f string) (*x509.Certificate, []byte, error) {
	in, err := os.ReadFile(f)
	if err != nil {
		return nil, nil, err
	}

	block, _ := pem.Decode(in)

	c, err := x509.ParseCertificate(block.Bytes)
	return c, in, err
}
