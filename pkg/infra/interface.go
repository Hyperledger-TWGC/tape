package infra

import (
	"tape/internal/fabric/protoutil"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
)

const (
	FULLPROCESS    = 6
	ENDORSEMENT    = 4
	COMMIT         = 3
	PROPOSALFILTER = 4
	COMMITFILTER   = 3
	QUERYFILTER    = 2
)

/*
to do for #127 SM crypto
just need to do an impl for this interface and replace
and impl a function for func (c Config) LoadCrypto() (*CryptoImpl, error) {
as generator
*/
type Crypto interface {
	protoutil.Signer
	NewSignatureHeader() (*common.SignatureHeader, error)
	/*Serialize() ([]byte, error)
	Sign(message []byte) ([]byte, error)*/
}

/*
as Tape major as Producer and Consumer pattern
define an interface here as Worker with start here
as for #56 and #174,in cli imp adjust sequence of P&C impl to control workflow.
*/
type Worker interface {
	Start()
}

type ObserverWorker interface {
	Worker
	GetTime() time.Time
}
