#!/bin/bash

echo
echo " ____    _____      _      ____    _____ "
echo "/ ___|  |_   _|    / \    |  _ \  |_   _|"
echo "\___ \    | |     / _ \   | |_) |   | |  "
echo " ___) |   | |    / ___ \  |  _ <    | |  "
echo "|____/    |_|   /_/   \_\ |_| \_\   |_|  "
echo
echo "Digital contracts signature, blockchain test network"
echo

CHANNEL_NAME="$1"
: ${CHANNEL_NAME:="digital-contracts-channel"}
: ${TIMEOUT:="60"}

COUNTER=1
MAX_RETRY=5
ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/berger-levrault.com/orderers/orderer.berger-levrault.com/msp/tlscacerts/tlsca.berger-levrault.com-cert.pem

echo "Channel name : "$CHANNEL_NAME

# verify the result of the end-to-end test
verifyResult () {
	if [ $1 -ne 0 ] ; then
		echo "!!!!!!!!!!!!!!! "$2" !!!!!!!!!!!!!!!!"
		echo "========= ERROR !!! FAILED to execute End-2-End Scenario ==========="
		echo
   		exit 1
	fi
}

setGlobals () {
	
	if [ $# -eq 2 ]; then	
		
		DOMAIN=$1
		MSPID=$2

		CORE_PEER_LOCALMSPID=$MSPID
		CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/$DOMAIN/peers/contract-service.$DOMAIN/tls/ca.crt
		CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/$DOMAIN/users/Admin@$DOMAIN/msp
		CORE_PEER_ADDRESS=contract-service.$DOMAIN:7051

		env | grep CORE
	else 
		echo "Bad syntax. Usage : setGlobals <DOMAIN> <MSPID>"
		echo
		exit 1
	fi
}

createChannel() {

	setGlobals berger-levrault.com BlMSP 

	if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
		peer channel create -o orderer.berger-levrault.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx >&log.txt
	else
		peer channel create -o orderer.berger-levrault.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA >&log.txt
	fi
	res=$?
	cat log.txt
	verifyResult $res "Channel creation failed"
	echo "===================== Channel \"$CHANNEL_NAME\" is created successfully ===================== "
	echo
}

# PARAMS :
# $1 = domain name of the peer being updated,
# $2 = MSP ID of the peer being updated
#
# ACTION :
#
# Update the specified anchor peer, acknowledging the chaincode changes on
# the channel.
#
updateAnchorPeers() {
	
	if [ $# -eq 2 ]; then

		DOMAIN=$1
		MSPID=$2

  		setGlobals $DOMAIN $MSPID 

  		if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
			peer channel update -o orderer.berger-levrault.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/${CORE_PEER_LOCALMSPID}anchors.tx >&log.txt
		else
			peer channel update -o orderer.berger-levrault.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/${CORE_PEER_LOCALMSPID}anchors.tx \
				--tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA >&log.txt
		fi
	
		res=$?
		cat log.txt
		verifyResult $res "Anchor peer update failed"
		echo "===================== Anchor peers for org \"$CORE_PEER_LOCALMSPID\" on \"$CHANNEL_NAME\" is updated successfully ===================== "
		echo
	else 
		echo "Bad syntax. Usage : updateAnchorPeers <DOMAIN> <MSPID>"
		echo
		exit 1
	fi
}

## Sometimes Join takes time hence RETRY atleast for 5 times
joinWithRetry () {
	
	peer channel join -b $CHANNEL_NAME.block  >&log.txt

	res=$?
	cat log.txt
	
	if [ $res -ne 0 -a $COUNTER -lt $MAX_RETRY ]; then
		COUNTER=` expr $COUNTER + 1`
		echo "$1 failed to join the channel, Retry after 2 seconds"
		sleep 2
		joinWithRetry $1
	else
		COUNTER=1
	fi
	
	verifyResult $res "After $MAX_RETRY attempts, PEER $1 has failed to Join the Channel"
}

joinChannel () {

	domain_names=(berger-levrault.com montpellier.fr)
	msp_ids=(BlMSP MtpMSP)
	
	for i in 0 1; do

		setGlobals ${domain_names[$i]} ${msp_ids[$i]}  
		joinWithRetry contract-service.${domain_names[$i]}
 
		echo "===================== contract-service.${domain_names[$i]} joined on the channel \"$CHANNEL_NAME\" ===================== "
		sleep 2
		echo

	done
}

installChaincode () {
	
	if [ $# -eq 3 ]; then
		
		DOMAIN=$1
		MSPID=$2
		CHAINCODE_NAME=$3
		
		setGlobals $DOMAIN $MSPID	
		peer chaincode install -n $CHAINCODE_NAME -v 1.0 -p github.com/hyperledger/fabric/peer/chaincode/go >&log.txt
		
		res=$?
		cat log.txt
			
		verifyResult $res "Chaincode installation on remote peer contract-service.$1 has Failed"
		echo "===================== Chaincode is installed on remote peer contract-service.$1 ===================== "
		echo
	else 
		echo "Bad syntax. Usage : installChaincode <DOMAIN> <MSPID> <CHAINCODE_NAME>"
		echo
		exit 1
	fi
}

instantiateChaincode () {
	
	if [ $# -eq 3 ]; then	
        	
		DOMAIN=$1
		MSPID=$2	
        	CHAINCODE_NAME=$3
	
		setGlobals $DOMAIN $MSPID
	
		# while 'peer chaincode' command can get the orderer endpoint from the peer (if join was successful),
		# lets supply it directly as we know it using the "-o" option
		if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
			peer chaincode instantiate -o orderer.berger-levrault.com:7050 -C $CHANNEL_NAME \
				-n $CHAINCODE_NAME -v 1.0 -c '{"Args":["a","100","b","200"]}' \
				-P "OR('BlMSP.admin','MtpMSP.admin')" >&log.txt
		else
			peer chaincode instantiate -o orderer.berger-levrault.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA \
				-C $CHANNEL_NAME -n $CHAINCODE_NAME -v 1.0 -c '{"Args":["a","100","b","200"]}' \
				-P "OR('BlMSP.admin', 'MtpMSP.admin')" >&log.txt
		fi

		res=$?
		cat log.txt
		verifyResult $res "Chaincode instantiation on PEER contract-service.$1 on channel '$CHANNEL_NAME' failed"
	
		echo "===================== Chaincode Instantiation on PEER contract-service.$1 on channel '$CHANNEL_NAME' is successful ===================== "
		echo
	else 
		echo "Bad syntax. Usage : instantiateChaincode <DOMAIN> <MSPID> <CHAINCODE_NAME>"
		echo
		exit 1
	fi
}

chaincodeQuery () {

  PEER=$1
  echo "===================== Querying on PEER$PEER on channel '$CHANNEL_NAME'... ===================== "
  setGlobals $PEER
  local rc=1
  local starttime=$(date +%s)

  # continue to poll
  # we either get a successful response, or reach TIMEOUT
  while test "$(($(date +%s)-starttime))" -lt "$TIMEOUT" -a $rc -ne 0
  do
     sleep 3
     echo "Attempting to Query PEER$PEER ...$(($(date +%s)-starttime)) secs"
     peer chaincode query -C $CHANNEL_NAME -n mycc -c '{"Args":["query","a"]}' >&log.txt
     test $? -eq 0 && VALUE=$(cat log.txt | awk '/Query Result/ {print $NF}')
     test "$VALUE" = "$2" && let rc=0
  done
  echo
  cat log.txt
  if test $rc -eq 0 ; then
	echo "===================== Query on PEER$PEER on channel '$CHANNEL_NAME' is successful ===================== "
  else
	echo "!!!!!!!!!!!!!!! Query result on PEER$PEER is INVALID !!!!!!!!!!!!!!!!"
        echo "================== ERROR !!! FAILED to execute End-2-End Scenario =================="
	echo
	exit 1
  fi
}

chaincodeInvoke () {
	PEER=$1
	setGlobals $PEER
	# while 'peer chaincode' command can get the orderer endpoint from the peer (if join was successful),
	# lets supply it directly as we know it using the "-o" option
	if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
		peer chaincode invoke -o orderer.berger-levrault.com:7050 -C $CHANNEL_NAME -n mycc -c '{"Args":["invoke","a","b","10"]}' >&log.txt
	else
		peer chaincode invoke -o orderer.berger-levrault.com:7050  --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n mycc -c '{"Args":["invoke","a","b","10"]}' >&log.txt
	fi
	res=$?
	cat log.txt
	verifyResult $res "Invoke execution on PEER$PEER failed "
	echo "===================== Invoke transaction on PEER$PEER on channel '$CHANNEL_NAME' is successful ===================== "
	echo
}

## Create channel
echo "Creating channel..."
createChannel

## Join all the peers to the channel
echo "Having all peers join the channel..."
joinChannel

## Set the anchor peers for each org in the channel
echo "Updating anchor peers for berger-levrault.com..."
updateAnchorPeers berger-levrault.com BlMSP 
echo "Updating anchor peers for montpellier.fr..."
updateAnchorPeers montpellier.fr MtpMSP

# Install chaincode on contract-service/montpellier.fr 
# and contract-service/berger-levrault.com
echo "Installing chaincode on contract-service.berger-levrault.com..."
installChaincode berger-levrault.com BlMSP digital-contracts-chaincode
echo "Install chaincode on contract-service.montpellier.fr..."
installChaincode montpellier.fr MtpMSP digital-contracts-chaincode

#Instantiate chaincode on contract-service.montpellier.fr 
echo "Instantiating chaincode on contract-service.montpellier.fr..."
instantiateChaincode montpellier.fr MtpMSP digital-contracts-chaincode

#Query on chaincode on Peer0/Org1
#echo "Querying chaincode on org1/peer0..."
#chaincodeQuery 0 100
#
##Invoke on chaincode on Peer0/Org1
#echo "Sending invoke transaction on org1/peer0..."
#chaincodeInvoke 0
#
## Install chaincode on Peer3/Org2
#echo "Installing chaincode on org2/peer3..."
#installChaincode 3
#
##Query on chaincode on Peer3/Org2, check if the result is 90
#echo "Querying chaincode on org2/peer3..."
#chaincodeQuery 3 90

echo
echo "========= All GOOD, digital contracts signature network updated =========== "
echo

echo
echo " _____   _   _   ____   "
echo "| ____| | \ | | |  _ \  "
echo "|  _|   |  \| | | | | | "
echo "| |___  | |\  | | |_| | "
echo "|_____| |_| \_| |____/  "
echo

exit 0
