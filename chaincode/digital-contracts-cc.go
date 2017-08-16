package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"time"
)
x
/* 
 * Implements the Chaincode interface
 */
type DigitalContractChaincode struct {
	
}

/*
 * The Init function is called at chaincode's instantiation time
 */
func (dc *DigitalContractChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

/*
 * Call to Invoke will generate a transaction proposal to write data to the ledger
 */
func (dc *DigitalContractChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	fun, args := stub.GetFunctionAndParameters()

	switch fun {
	case "initLedger":
		return dc.initLedger(stub)
		
	case "addContract":
		return dc.addContract(stub, args)
		
	case "getContract":
		return dc.getContract(stub, args)
	default:
		return shim.Error(fmt.Sprintf("Trying to invoke unknown function %s", fun))
	}

}

/*
 * Function to initialize the ledger with a first asset.
 * First asset key is 0, and value is the byte array of a marshaled JSON contract.
 */
func (dc *DigitalContractChaincode) initLedger(stub shim.ChaincodeStubInterface) peer.Response {

	firstContract, err := stub.GetState("0")

	/* Check if contract with 0 exists*/
	if firstContract != nil {
		return shim.Error("Ledger already initialized")
	} else if err != nil {
		return shim.Error(err.Error())
	}
		
	/* Creating signatories for the contract */
	client := Signatory{
		BusinessName: "Ville de Montpellier",
		HeadQuarters: "1, Place Georges Frêche, 34000 Montpellier",
		Holder: "Philippe Saurel",
		RegistrationNumber: "213 401 722"}
	
	contractor := Signatory{
		BusinessName: "Berger-Levrault",
		HeadQuarters: "892, Rue Yves Kermen, 92100 Boulogne-Billancourt",
		Holder: "Antoine Rouillard",
		RegistrationNumber: "755 800 646"}

	/* Creating Contract object */
	contract := Contract{
		Signatures: []ContractSignature{},
		ContractHeading: "Maintenance gestion des ressources humaines et gestion financière",
		StartingDate: time.Now(),
		EndingDate: time.Time{},
		StateRecords: []ContractState{},
		PaymentRecords: []Payment{}}
	
	// Declaring firstState outside contract initialization
	// for cross-referencing
	genesisState := ContractState{
		Heading: WAITING_FOR_SIGNATURE,
		StartingDate: time.Now(),
		EndingDate: time.Time{}}

	contract.StateRecords = append(contract.StateRecords, genesisState)

	// Signatures issued by client and contractor
	clientSignature := ContractSignature{
		SignatoryRef: &client,
		StatusOfSignatory: CLIENT,
		DateOfSignature: time.Now(),
		SignatureDigest: fmt.Sprintf("%x", sha256.Sum256([]byte(client.RegistrationNumber)))}

	contractorSignature := ContractSignature{
		SignatoryRef: &contractor,
		StatusOfSignatory: CONTRACTOR,
		DateOfSignature: time.Now(),
		SignatureDigest: fmt.Sprintf("%x", sha256.Sum256([]byte(contractor.RegistrationNumber)))}

	// Add signatures to the Signatures slice in the contract object
	contract.Signatures = append(contract.Signatures, clientSignature)
	contract.Signatures = append(contract.Signatures, contractorSignature)
	
	// Storing Contract object in the ledger with key 0
	marshaledContract, err := json.Marshal(contract)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState("0", marshaledContract)
	if err != nil {
		return shim.Error("Failed to create genesis asset")
	}

	fmt.Println(contract)
	fmt.Println(string(marshaledContract))
	
	return shim.Success(marshaledContract)
}

/*
 * Function returning the contract object from the given key if exists,
 * else returns an error message.
 */
func (dc *DigitalContractChaincode) getContract(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	
	if len(args) > 0 {

		contractKey := args[0]
		contractBytes, err := stub.GetState(contractKey)

		if err != nil {
			return shim.Error(err.Error())
		} else if contractBytes == nil {
			return shim.Error(fmt.Sprintf("Asset with key %v doesn't exist", contractKey))
		} else {
			contract := new(Contract)
			json.Unmarshal(contractBytes, contract)

			fmt.Println(*contract)
			return shim.Success(contractBytes)
		}
		
	} else {
		return shim.Error(fmt.Sprintf("Wrong number of arguments. Given %v expected 1 (a key)", len(args)))
	}
	
}

/*
 * Function to add a contract in the ledger.
 * New asset key must be a unique number (corresponding to the contract id of the relational DB),
 * and new asset value must be a JSON object representing a contract struct such as defined
 * in the `digital-contracts-structs.go` file. 
 */
func (dc *DigitalContractChaincode) addContract(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	// args not empty and filled with a key-value pair
	if (len(args) == 2) {

		newContractKey, newContract := args[0], args[1]
		
		existingContract, err := stub.GetState(newContractKey)
		if err != nil {
			
			return shim.Error(err.Error())
			
		} else if existingContract != nil {
			
			return shim.Error(fmt.Sprintf("Asset %s already exists", newContractKey))
			
		} else {
			// We store the key and the value on the ledger
			err = stub.PutState(newContractKey, []byte(newContract))
			
			if err != nil {
				return shim.Error(fmt.Sprintf("Failed to create asset: %s with value: %s", newContractKey, newContract))
			}
		}

		return shim.Success([]byte(newContract))

		
	} else {
		
		return shim.Error("Wrong number of arguments. Expecting 2 arguments (key and value)")

	}

}

// main function starts up the chaincode in the container during instantiate
func main() {
	
	if err := shim.Start(new(DigitalContractChaincode)); err != nil {
		fmt.Printf("Error starting DigitalContractChaincode chaincode: %s", err)
	}
}
