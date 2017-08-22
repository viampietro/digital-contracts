package main

import (
	"strconv"
)

/*=======================================================
 =============== CONTRACT STATE HEADING =================
 ========================================================
        Defining the contract's state as enum
 ========================================================*/
type ContractStateHeading int

const (
	WAITING_FOR_SIGNATURE = ContractStateHeading(iota)
	SIGNED
	WAITING_FOR_PAYMENT
	IN_ORDER
)

/*
 * String method is called whenever trying to print a ContractStateHeading object
 */
func (csh ContractStateHeading) String() string {
	
	name := []string{"en attente de signature", "signé", "en attente de paiement", "en règle"}
	i := int(csh)

	switch {
	case i <= int(IN_ORDER):
		return name[i]
	default:
		return strconv.Itoa(int(i))
	}
}

/*===============================================
 =============== SIGNATORY STATUS ===============
 ================================================
     Defining the signatory's status as enum
 ================================================*/
type SignatoryStatus int

const (
	CLIENT = SignatoryStatus(iota)
	CONTRACTOR
)

/*
 * String method is called whenever trying to print a SignatoryStatus object
 */
func (sigstat SignatoryStatus) String() string {
	
	name := []string{"client", "prestataire"}
	i := int(sigstat)

	switch {
	case i <= int(CONTRACTOR):
		return name[i]
	default:
		return strconv.Itoa(int(i))
	}
}

/*===============================================
 =============== CONTRACT STATE =================
 ================================================
  Defining struct for contract's state
 ================================================*/
type ContractState struct {
	
	Heading      ContractStateHeading // ContractState's description  
	StartingDate string               // Starting date for ContractState
	EndingDate   string               // Ending date for ContractState (nil for current state)

}

/*======================================
 =============== PAYMENT ===============
 =======================================
 Defining struct for payment's issuance
 =======================================*/
type Payment struct {

	Amount         float64    // payment's amount in euros
	DateOfIssuance string     // payment's date
	Issuer         Signatory  // ref to signatory issuing the payment
	
}

/*=========================================
 =============== SIGNATORY ================
 ==========================================
 Defining struct for contract's signatories 
 ==========================================*/
type Signatory struct {

	BusinessName       string          // Name of legal entity          
	HeadQuarters       string          // Address of the signatory's head quarters 
	Holder             string          // Name of legal holder
	RegistrationNumber string          // SIRET number
	Status             SignatoryStatus // Signatory status relative to the contract
}

/*=================================================
 =============== CONTRACTSIGNATURE ================
 ==================================================
     Defining struct for contract's signatures 
 ==================================================*/
type ContractSignature struct {

	Issuer            Signatory // Signatory issuing this signature
	DateOfSignature   string    // Contract's date of signature
	SignatureDigest   string    // SHA256 digest of signatory's private signature
}

/*=================================================
 ==================== CONTRACT ====================
 ==================================================
    Defining struct for contracts, 
    which will be stored as values in the ledger 
 ==================================================*/
type Contract struct {
	
	ContractHeading string              // Object of the contract
	StartingDate    string              // Contract's starting date
	EndingDate      string              // Contract's ending date
	StateRecords    []ContractState     // History of contract's previous state
	PaymentRecords  []Payment           // Payment issued in regard to the contract's fulfillment
	Signatories     []Signatory         // Contract's parties
	Signatures      []ContractSignature // All contract's signatures

}
