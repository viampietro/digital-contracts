package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// Implements the Chaincode interface
type DigitalContractChaincode struct {

}

func (dc *DigitalContractChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {

	args := stub.GetStringArgs()

	// args not empty and filled with key-value pairs
	if (len(args) > 0) && (len(args) % 2 == 0) {

		// loop through all key-value entries and check
		// if key doesn't exist in state already
		for i := 0; i < len(args); i += 2 {

			value, err := stub.GetState(args[i])
			if err != nil {
				
				return shim.Error(err.Error())
				
			} else if value == nil {

				// We store the key and the value on the ledger
				err = stub.PutState(args[i], []byte(args[i+1]))
				if err != nil {
					return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[i]))
				}
			}
		}
		
	} else {
		return shim.Error("Incorrect arguments. Expecting at least a key and a value")
	}

	return shim.Success(nil)
}

func (dc *DigitalContractChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// main function starts up the chaincode in the container during instantiate
func main() {
    if err := shim.Start(new(DigitalContractChaincode)); err != nil {
            fmt.Printf("Error starting DigitalContractChaincode chaincode: %s", err)
    }
}
