package infra

import (
	"bytes"
	"math"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
)

func CreateProposal(signer *Crypto, channel, ccname string, args ...string) *peer.Proposal {
	var argsInByte [][]byte
	for _, arg := range args {
		argsInByte = append(argsInByte, []byte(arg))
	}

	spec := &peer.ChaincodeSpec{
		Type:        peer.ChaincodeSpec_GOLANG,
		ChaincodeId: &peer.ChaincodeID{Name: ccname},
		Input:       &peer.ChaincodeInput{Args: argsInByte},
	}

	invocation := &peer.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	creator, err := signer.Serialize()
	if err != nil {
		panic(err)
	}

	prop, _, err := utils.CreateChaincodeProposal(common.HeaderType_ENDORSER_TRANSACTION, channel, invocation, creator)
	if err != nil {
		panic(err)
	}

	return prop
}

func SignProposal(prop *peer.Proposal, signer *Crypto) (*peer.SignedProposal, error) {
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

func CreateSignedTx(proposal *peer.Proposal, signer *Crypto, resps ...*peer.ProposalResponse) (*common.Envelope, error) {
	if len(resps) == 0 {
		return nil, errors.New("at least one proposal response is required")
	}

	// the original header
	hdr, err := utils.GetHeader(proposal.Header)
	if err != nil {
		return nil, err
	}

	// the original payload
	pPayl, err := utils.GetChaincodeProposalPayload(proposal.Payload)
	if err != nil {
		return nil, err
	}

	// check that the signer is the same that is referenced in the header
	// TODO: maybe worth removing?
	signerBytes, err := signer.Serialize()
	if err != nil {
		return nil, err
	}

	shdr, err := utils.GetSignatureHeader(hdr.SignatureHeader)
	if err != nil {
		return nil, err
	}

	if bytes.Compare(signerBytes, shdr.Creator) != 0 {
		return nil, errors.New("signer must be the same as the one referenced in the header")
	}

	// get header extensions so we have the visibility field
	hdrExt, err := utils.GetChaincodeHeaderExtension(hdr)
	if err != nil {
		return nil, err
	}

	// ensure that all actions are bitwise equal and that they are successful
	var a1 []byte
	for n, r := range resps {
		if n == 0 {
			a1 = r.Payload
			if r.Response.Status < 200 || r.Response.Status >= 400 {
				return nil, errors.Errorf("proposal response was not successful, error code %d, msg %s", r.Response.Status, r.Response.Message)
			}
			continue
		}

		if bytes.Compare(a1, r.Payload) != 0 {
			return nil, errors.New("ProposalResponsePayloads do not match")
		}
	}

	// fill endorsements
	endorsements := make([]*peer.Endorsement, len(resps))
	for n, r := range resps {
		endorsements[n] = r.Endorsement
	}

	// create ChaincodeEndorsedAction
	cea := &peer.ChaincodeEndorsedAction{ProposalResponsePayload: resps[0].Payload, Endorsements: endorsements}

	// obtain the bytes of the proposal payload that will go to the transaction
	propPayloadBytes, err := utils.GetBytesProposalPayloadForTx(pPayl, hdrExt.PayloadVisibility)
	if err != nil {
		return nil, err
	}

	// serialize the chaincode action payload
	cap := &peer.ChaincodeActionPayload{ChaincodeProposalPayload: propPayloadBytes, Action: cea}
	capBytes, err := utils.GetBytesChaincodeActionPayload(cap)
	if err != nil {
		return nil, err
	}

	// create a transaction
	taa := &peer.TransactionAction{Header: hdr.SignatureHeader, Payload: capBytes}
	taas := make([]*peer.TransactionAction, 1)
	taas[0] = taa
	tx := &peer.Transaction{Actions: taas}

	// serialize the tx
	txBytes, err := utils.GetBytesTransaction(tx)
	if err != nil {
		return nil, err
	}

	// create the payload
	payl := &common.Payload{Header: hdr, Data: txBytes}
	paylBytes, err := utils.GetBytesPayload(payl)
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

func CreateSignedDeliverNewestEnv(ch string, signer *Crypto) (*common.Envelope, error) {
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

	return utils.CreateSignedEnvelope(
		common.HeaderType_DELIVER_SEEK_INFO,
		ch,
		signer,
		seekInfo,
		0,
		0,
	)
}
