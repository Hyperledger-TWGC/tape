package main

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// PutGet example simple Chaincode implementation
type PutGet struct {
	contractapi.Contract
}

func (t *PutGet) Init(ctx contractapi.TransactionContextInterface) error {
	return nil
}

func (t *PutGet) Put(ctx contractapi.TransactionContextInterface, key, value string) error {
	err := ctx.GetStub().PutState(key, []byte(value))
	if err != nil {
		return fmt.Errorf("failed to store k-v: %s", err)
	}

	return nil
}

func (t *PutGet) Get(ctx contractapi.TransactionContextInterface, key string) (string, error) {
	res, err := ctx.GetStub().GetState(key)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

func main() {
	cc, err := contractapi.NewChaincode(new(PutGet))
	if err != nil {
		panic(err.Error())
	}
	if err := cc.Start(); err != nil {
		fmt.Printf("Error starting ABstore chaincode: %s", err)
	}
}
