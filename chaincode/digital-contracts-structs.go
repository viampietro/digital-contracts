package main

import (
	"strconv"
	"time"
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
	
	name := []string{"waiting for signature", "signed", "waiting for payment", "in order"}
	i := int(csh)

	switch {
	case i <= int(IN_ORDER):
		return name[i]
	default:
		return strconv.Itoa(int(i))
	}
}

/*==============================================
 =============== SIGNATORYSTATUS ===============
 ===============================================
     Defining the signatory's status as enum
 ===============================================*/
type SignatoryStatus int

const (
	CLIENT = SignatoryStatus(iota)
	CONTRACTOR
)

/*
 * String method is called whenever trying to print a SignatoryStatus object
 */
func (sigstat SignatoryStatus) String() string {
	
	name := []string{"client", "contractor"}
	i := int(sigstat)

	switch {
	case i <= int(CONTRACTOR):
		return name[i]
	default:
		return strconv.Itoa(int(i))
	}
}

/*==============================================
 =============== CONTRACTSTATE =================
 ===============================================
  Defining struct to express the contract state
 ===============================================*/
type ContractState struct {
	
	Heading      ContractStateHeading // ContractState's description  
	StartingDate time.Time            // Starting date for ContractState
	EndingDate   time.Time            // Ending date for ContractState (nil for current state)

}

/*======================================
 =============== PAYMENT ===============
 =======================================
 Defining struct for payment's issuance
 =======================================*/
type Payment struct {

	Amount         float64    // payment's amount in euros
	DateOfIssuance time.Time  // payment's date
	Issuer         *Signatory // ref to signatory issuing the payment
	
}

/*=========================================
 =============== SIGNATORY ================
 ==========================================
 Defining struct for contract's signatories 
 ==========================================*/
type Signatory struct {

	BusinessName       string // Name of legal entity          
	HeadQuarters       string // Address of the signatory's head quarters 
	Holder             string // Name of legal holder
	RegistrationNumber string // SIRET number

}

/*=================================================
 =============== CONTRACTSIGNATURE ================
 ==================================================
     Defining struct for contract's signatures 
 ==================================================*/
type ContractSignature struct {

	SignatoryRef      *Signatory      // Ref to signatory issuing this signature
	StatusOfSignatory SignatoryStatus // Signatory status relative to the contract
	DateOfSignature   time.Time       // Contract's date of signature
	SignatureDigest   string          // SHA256 digest of signatory's private signature
}

/*=================================================
 ==================== CONTRACT ====================
 ==================================================
    Defining struct for contracts, 
    which will be stored as values in the ledger 
 ==================================================*/
type Contract struct {
	
	Signatures      []ContractSignature // All contract's signatures
	ContractHeading string              // Object of the contract
	StartingDate    time.Time           // Contract's starting date
	EndingDate      time.Time           // Contract's ending date
	StateRecords    []ContractState     // History of contract's previous state
	PaymentRecords  []Payment           // Payment issued in regard to the contract's fulfillment
}
