'use strict';

/*
 * Hyperledger Fabric Sample Query Program
 */

var hfc = require('fabric-client');
var path = require('path');

var options = {

  wallet_path: path.join(__dirname, './creds'),
  private_key_path: path.join(__dirname, './crypto-config/peerOrganizations/montpellier.fr/users/Admin@montpellier.fr/msp/keystore/2f6b11ce722ce150d27946dae3c6c0bddc989f65c6246924ece0595e14777c04_sk'),
  ecert_path: path.join(__dirname, './crypto-config/peerOrganizations/montpellier.fr/users/Admin@montpellier.fr/msp/signcerts/Admin@montpellier.fr-cert.pem'),
  user_id: 'Vincent',
  channel_id: 'digital-contracts-channel',
  chaincode_id: 'digital-contracts-chaincode',
  network_url: 'grpc://contract-service.montpellier.fr:7051'

};

var channel = {};
var client = null;

Promise.resolve().then(() => {

  console.log("Create a client and set the wallet location");
  client = new hfc();

  return hfc.newDefaultKeyValueStore({ path: options.wallet_path });

}).then((wallet) => {

  console.log("Set wallet path, and associate user ", options.user_id, " with application");
  client.setStateStore(wallet);

  return client.createUser({username: "Vincent",
                            mspid: "MtpMSP",
                            cryptoContent: {privateKey: options.private_key_path,
                                            signedCert: options.ecert_path}});

}).then((user) => {

  console.log("Check user is enrolled, and set a query URL in the network");

  if (user === null || user.isEnrolled() === false) {
    console.error("User not defined, or not enrolled - error");
  }

  channel = client.newChannel(options.channel_id);
  channel.addPeer(client.newPeer(options.network_url));
  
  return;

}).then(() => {

  console.log("Make query");

  var transaction_id = client.newTransactionID();

  console.log("Assigning transaction_id: ", transaction_id._transaction_id);

  // queryCar - requires 1 argument, ex: args: ['CAR4'],
  // queryAllCars - requires no arguments , ex: args: [''],
  const request = {
    chaincodeId: options.chaincode_id,
    txId: transaction_id,
    fcn: 'initLedger',
    args: ['']
  };

  return channel.queryByChaincode(request);

}).then((query_responses) => {

  console.log("returned from query");

  if (!query_responses.length) {
    console.log("No payloads were returned from query");
  } else {
    console.log("Query result count = ", query_responses.length)
  }
  
  if (query_responses[0] instanceof Error) {
    console.error("error from query = ", query_responses[0]);
  }
  
  console.log("Response is ", query_responses[0].toString());

}).catch((err) => {

  console.error("Caught Error", err);

});
