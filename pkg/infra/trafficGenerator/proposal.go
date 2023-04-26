package trafficGenerator

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hyperledger-twgc/tape/internal/fabric/protoutil"
	"github.com/hyperledger-twgc/tape/pkg/infra"
	"github.com/hyperledger-twgc/tape/pkg/infra/basic"
	"github.com/opentracing/opentracing-go"

	"github.com/google/uuid"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func CreateProposal(signer infra.Crypto, logger *log.Logger, channel, ccname, version string, args ...string) (*basic.TracingProposal, error) {
	var argsInByte [][]byte
	for _, arg := range args {
		current_arg, err := ConvertString(arg)
		if err != nil {
			return nil, err
		}
		fmt.Println(current_arg)
		argsInByte = append(argsInByte, []byte(current_arg))
	}

	spec := &peer.ChaincodeSpec{
		Type:        peer.ChaincodeSpec_GOLANG,
		ChaincodeId: &peer.ChaincodeID{Name: ccname, Version: version},
		Input:       &peer.ChaincodeInput{Args: argsInByte},
	}

	invocation := &peer.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	creator, err := signer.Serialize()
	if err != nil {
		return nil, err
	}

	prop, txid, err := protoutil.CreateChaincodeProposal(common.HeaderType_ENDORSER_TRANSACTION, channel, invocation, creator)
	if err != nil {
		return nil, err
	}
	basic.LogEvent(logger, txid, "CreateChaincodeProposal")
	tapeSpan := basic.GetGlobalSpan()
	var span opentracing.Span
	if basic.GetMod() == infra.FULLPROCESS {
		Global_Span := tapeSpan.SpanIntoMap(txid, "", basic.TRANSCATION, nil)
		span = tapeSpan.MakeSpan(txid, "", basic.TRANSCATIONSTART, Global_Span)
	} else {
		span = tapeSpan.SpanIntoMap(txid, "", basic.TRANSCATIONSTART, nil)
	}
	return &basic.TracingProposal{Proposal: prop, TxId: txid, Span: span}, nil
}

func SignProposal(prop *peer.Proposal, signer infra.Crypto) (*peer.SignedProposal, error) {
	propBytes, err := proto.Marshal(prop)
	if err != nil {
		return nil, err
	}

	sig, err := signer.Sign(propBytes)
	if err != nil {
		return nil, err
	}

	return &peer.SignedProposal{ProposalBytes: propBytes, Signature: sig}, nil
}

func CreateSignedTx(signedproposal *peer.SignedProposal, signer infra.Crypto, resps []*peer.ProposalResponse) (*common.Envelope, error) {
	if len(resps) == 0 {
		return nil, errors.Errorf("at least one proposal response is required")
	}
	proposal := &peer.Proposal{}
	err := proto.Unmarshal(signedproposal.ProposalBytes, proposal)
	if err != nil {
		return nil, err
	}
	// the original header
	hdr, err := GetHeader(proposal.Header)
	if err != nil {
		return nil, err
	}

	// the original payload
	pPayl, err := GetChaincodeProposalPayload(proposal.Payload)
	if err != nil {
		return nil, err
	}

	// check that the signer is the same that is referenced in the header
	// TODO: maybe worth removing?
	signerBytes, err := signer.Serialize()
	if err != nil {
		return nil, err
	}

	shdr, err := GetSignatureHeader(hdr.SignatureHeader)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(signerBytes, shdr.Creator) {
		return nil, errors.Errorf("signer must be the same as the one referenced in the header")
	}

	// get header extensions so we have the visibility field
	_, err = GetChaincodeHeaderExtension(hdr)
	if err != nil {
		return nil, err
	}

	endorsements := make([]*peer.Endorsement, 0)

	// ensure that all actions are bitwise equal and that they are successful
	var a1 []byte
	for n, r := range resps {
		if n == 0 {
			a1 = r.Payload
			if r.Response.Status < 200 || r.Response.Status >= 400 {
				return nil, errors.Errorf("proposal response was not successful, error code %d, msg %s", r.Response.Status, r.Response.Message)
			}
		}
		if !bytes.Equal(a1, r.Payload) {
			return nil, errors.Errorf("ProposalResponsePayloads from Peers do not match")
		}
		endorsements = append(endorsements, r.Endorsement)
	}
	// create ChaincodeEndorsedAction
	cea := &peer.ChaincodeEndorsedAction{ProposalResponsePayload: a1, Endorsements: endorsements}

	// obtain the bytes of the proposal payload that will go to the transaction
	propPayloadBytes, err := protoutil.GetBytesProposalPayloadForTx(pPayl) //, hdrExt.PayloadVisibility
	if err != nil {
		return nil, err
	}

	// serialize the chaincode action payload
	cap := &peer.ChaincodeActionPayload{ChaincodeProposalPayload: propPayloadBytes, Action: cea}
	capBytes, err := protoutil.GetBytesChaincodeActionPayload(cap)
	if err != nil {
		return nil, err
	}

	// create a transaction
	taa := &peer.TransactionAction{Header: hdr.SignatureHeader, Payload: capBytes}
	taas := make([]*peer.TransactionAction, 1)
	taas[0] = taa
	tx := &peer.Transaction{Actions: taas}
	// serialize the tx
	txBytes, err := protoutil.GetBytesTransaction(tx)
	if err != nil {
		return nil, err
	}

	// create the payload
	payl := &common.Payload{Header: hdr, Data: txBytes}
	paylBytes, err := protoutil.GetBytesPayload(payl)
	if err != nil {
		return nil, err
	}

	// sign the payload
	sig, err := signer.Sign(paylBytes)
	if err != nil {
		return nil, err
	}
	// here's the envelope
	return &common.Envelope{Payload: paylBytes, Signature: sig}, nil
}

func CreateSignedDeliverNewestEnv(ch string, signer infra.Crypto) (*common.Envelope, error) {
	start := &orderer.SeekPosition{
		Type: &orderer.SeekPosition_Newest{
			Newest: &orderer.SeekNewest{},
		},
	}

	stop := &orderer.SeekPosition{
		Type: &orderer.SeekPosition_Specified{
			Specified: &orderer.SeekSpecified{
				Number: math.MaxUint64,
			},
		},
	}

	seekInfo := &orderer.SeekInfo{
		Start:    start,
		Stop:     stop,
		Behavior: orderer.SeekInfo_BLOCK_UNTIL_READY,
	}

	return protoutil.CreateSignedEnvelope(
		common.HeaderType_DELIVER_SEEK_INFO,
		ch,
		signer,
		seekInfo,
		0,
		0,
	)
}

func GetHeader(bytes []byte) (*common.Header, error) {
	hdr := &common.Header{}
	err := proto.Unmarshal(bytes, hdr)
	return hdr, errors.Wrap(err, "error unmarshaling Header")
}

func GetChaincodeProposalPayload(bytes []byte) (*peer.ChaincodeProposalPayload, error) {
	cpp := &peer.ChaincodeProposalPayload{}
	err := proto.Unmarshal(bytes, cpp)
	return cpp, errors.Wrap(err, "error unmarshaling ChaincodeProposalPayload")
}

func GetSignatureHeader(bytes []byte) (*common.SignatureHeader, error) {
	return UnmarshalSignatureHeader(bytes)
}

func GetChaincodeHeaderExtension(hdr *common.Header) (*peer.ChaincodeHeaderExtension, error) {
	chdr, err := UnmarshalChannelHeader(hdr.ChannelHeader)
	if err != nil {
		return nil, err
	}

	chaincodeHdrExt := &peer.ChaincodeHeaderExtension{}
	err = proto.Unmarshal(chdr.Extension, chaincodeHdrExt)
	return chaincodeHdrExt, errors.Wrap(err, "error unmarshaling ChaincodeHeaderExtension")
}

// UnmarshalChannelHeader returns a ChannelHeader from bytes
func UnmarshalChannelHeader(bytes []byte) (*common.ChannelHeader, error) {
	chdr := &common.ChannelHeader{}
	err := proto.Unmarshal(bytes, chdr)
	return chdr, errors.Wrap(err, "error unmarshaling ChannelHeader")
}

func UnmarshalSignatureHeader(bytes []byte) (*common.SignatureHeader, error) {
	sh := &common.SignatureHeader{}
	if err := proto.Unmarshal(bytes, sh); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling SignatureHeader")
	}
	return sh, nil
}

func newUUID() string {
	newUUID, _ := uuid.NewRandom()
	return newUUID.String()
}

func randomInt(min, max int) int {
	if min < 0 {
		min = 0
	}

	if max <= 0 {
		max = 1
	}

	return seededRand.Intn(max-min) + min
}

const maxLen = 16
const minLen = 2

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func randomString(length int) string {
	if length <= 0 {
		length = seededRand.Intn(maxLen-minLen+1) + minLen
	}

	return stringWithCharset(length, charset)
}

func ConvertString(arg string) (string, error) {
	// ref to https://ghz.sh/docs/calldata
	// currently supports three kinds of random
	// support for uuid
	// uuid
	// support for random strings
	// randomString$length
	// support for random int
	// randomNumberMin_Max
	if !utf8.ValidString(arg) {
		return "", errors.New("invalid string")
	}
	var current_arg = arg
	regUUID, _ := regexp.Compile("uuid")
	//FindAllStringIndex
	// if reg.FindAllStringIndex !=nil
	// i=0;i<len;i=i+2
	// cal value
	// replace 1
	finds := regUUID.FindAllStringIndex(current_arg, -1)
	for _, v := range finds {
		str := fmt.Sprint(arg[v[0]:v[1]])
		current_arg = strings.Replace(current_arg, str, newUUID(), 1)
	}
	regString, _ := regexp.Compile("randomString(\\d*)")
	finds = regString.FindAllStringIndex(current_arg, -1)
	arg = current_arg
	for _, v := range finds {
		str := fmt.Sprint(arg[v[0]:v[1]])
		length, err := strconv.Atoi(strings.TrimPrefix(str, "randomString"))
		if err != nil {
			return arg, err
		}
		if length > 4096 {
			return arg, fmt.Errorf("random string over length of 4096")
		}
		current_arg = strings.Replace(current_arg, str, randomString(length), 1)
	}
	regNumber, _ := regexp.Compile("randomNumber(\\d*)_(\\d*)")
	arg = current_arg
	finds = regNumber.FindAllStringIndex(current_arg, -1)
	for _, v := range finds {
		str := fmt.Sprint(arg[v[0]:v[1]])
		min_maxStr := strings.TrimPrefix(str, "randomNumber")
		min_maxArray := strings.Split(min_maxStr, "_")
		min, err := strconv.Atoi(min_maxArray[0])
		if err != nil {
			return arg, err
		}
		max, err := strconv.Atoi(min_maxArray[1])
		if err != nil {
			return arg, err
		}
		if max <= min {
			return arg, fmt.Errorf("max less than min, or equal")
		}
		current_arg = strings.Replace(current_arg, str, strconv.Itoa(randomInt(min, max)), 1)
	}
	return current_arg, nil
}
