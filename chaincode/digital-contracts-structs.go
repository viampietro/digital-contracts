package main

import (
	"strconv"
	"time"
)

/*
 * Defining the contract's different state as enum
 */
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

/*
 * Defining struct to express the contract state
 * 
 */
type ContractState struct {
	
	Heading      ContractStateHeading // ContractState's description  
	StartingDate time.Time                 // Starting date for ContractState
	EndingDate   time.Time                 // Ending date for ContractState (nil for current state)

}

/*
 * 
 *  Defining struct for payment's issuance
 * 
 */
type Payment struct {

	Amount         float64    // payment's amount in euros
	DateOfIssuance time.Time  // payment's date
	Issuer         *Signatory // ref to signatory issuing the payment
	
}


/*
 * Defining struct for contract's signatories 
 *
 */
type Signatory struct {

	BusinessName       string // Name of legal entity          
	HeadQuarters       string // Address of the signatory's head quarters 
	Holder             string // Name of legal holder
	RegistrationNumber string // SIRET number

}

/*
 * Defining struct for contracts, 
 * which will be stored as values in the ledger 
 *
 */
type Contract struct {
	
	Client          *Signatory      // Ref to signatory representing the client side 
	Contractor      *Signatory      // Ref to signatory representing the contractor side
	ContractHeading string          // Object of the contract
	StartingDate    time.Time            // Contract's starting date
	EndingDate      time.Time            // Contract's ending date
	StateRecords    []ContractState // History of contract's state
	PaymentRecords  []Payment       // Payment issued in regard to the contract's fulfillment
}
