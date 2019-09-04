/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// PutGet example simple Chaincode implementation
type PutGet struct {
}

func (t *PutGet) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *PutGet) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case "put":
		return t.put(stub, args)
	case "get":
		return t.get(stub, args)
	default:
		return shim.Error(fmt.Sprintf("unknown func: %s", function))
	}
}

func (t *PutGet) put(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of args, expecting 2")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(args[0]))
}

func (t *PutGet) get(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of args, expecting 1")
	}

	res, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(res)
}

func main() {
	err := shim.Start(new(PutGet))
	if err != nil {
		fmt.Printf("Error starting PutGet chaincode: %s", err)
	}
}
