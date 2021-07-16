package infra

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/pem"
	"github.com/tjfoc/gmsm/sm3"
	"io/ioutil"
	"math/big"

	"tape/internal/fabric/bccsp/utils"
	"tape/internal/fabric/common/crypto"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/pkg/errors"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
)

type CryptoConfig struct {
	MSPID      string
	PrivKey    string
	SignCert   string
	TLSCACerts []string
}

type ECDSASignature struct {
	R, S *big.Int
}

type Crypto struct {
	Creator  []byte
	PrivKey  *sm2.PrivateKey
	SignCert *x509.Certificate
}

func (s *Crypto) Sign(message []byte) ([]byte, error) {
	sign, err := s.PrivKey.Sign(rand.Reader, digest(message), nil)
	if err != nil {
		return nil, err
	}
	return sign, nil
}

func (s *Crypto) Serialize() ([]byte, error) {
	return s.Creator, nil
}

func (s *Crypto) NewSignatureHeader() (*common.SignatureHeader, error) {
	creator, err := s.Serialize()
	if err != nil {
		return nil, err
	}
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return nil, err
	}

	return &common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
	}, nil
}

func digest(in []byte) []byte {
	h := sm3.New()
	h.Write(in)
	return h.Sum(nil)
}

func toPEM(in []byte) ([]byte, error) {
	d := make([]byte, base64.StdEncoding.DecodedLen(len(in)))
	n, err := base64.StdEncoding.Decode(d, in)
	if err != nil {
		return nil, err
	}
	return d[:n], nil
}

func GetPrivateKey(f string) (*sm2.PrivateKey, error) {
	in, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}

	k, err := utils.PEMtoPrivateKey(in, []byte{})
	if err != nil {
		return nil, err
	}

	key, ok := k.(*sm2.PrivateKey)
	if !ok {
		return nil, errors.Errorf("expecting ecdsa key")
	}

	return key, nil
}

func GetCertificate(f string) (*x509.Certificate, []byte, error) {
	in, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, nil, err
	}

	block, _ := pem.Decode(in)

	c, err := x509.ParseCertificate(block.Bytes)
	return c, in, err
}
