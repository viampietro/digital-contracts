'use strict';

var express = require('express');
var app = express();
var fs = require("fs");
var hfc = require('fabric-client');
var path = require('path');
var util = require('util');
var exec = require('child_process').exec;
var bodyParser = require('body-parser');

// configure the app to use bodyParser()
app.use(bodyParser.urlencoded({
  extended: true
}));

app.use(bodyParser.json()); // for parsing application/json

/*******************************************
 * NETWORK AND CRYPTO-MATERIAL INFORMATION *
 *******************************************/
var path_to_user_crypto_material = '../crypto-config/peerOrganizations/berger-levrault.com/users/Admin@berger-levrault.com/msp';

// the filename of the user's private key change every time the network's crypto-material is reset
// but the file's location is always the same
var private_key_promise = new Promise ((resolve, reject) => {
  exec('ls ' + path.join(__dirname, path_to_user_crypto_material, 'keystore'), (error, stdout, stderr) => {
    if (error) {
      reject(`child_process.exec error : ${error}`);
      return;
    }
    resolve(stdout.trim());
  });
});

var options = {

  wallet_path: path.join(__dirname, 'creds'),
  private_key_path: path.join(__dirname, path_to_user_crypto_material, 'keystore'),
  ecert_path: path.join(__dirname, path_to_user_crypto_material, 'signcerts/Admin@berger-levrault.com-cert.pem'),
  user_id: 'BLOperator',
  mspid: 'BlMSP',
  channel_id: 'digital-contracts-channel',
  chaincode_id: 'digital-contracts-chaincode',
  peer_url: 'grpc://localhost:7051',
  event_url: 'grpc://localhost:7053',
  orderer_url: 'grpc://localhost:7050'

};

var channel = {};
var client = null;
var targets = [];
var tx_id = null;

/****************************************************************************
 ****************************************************************************  
 ***************** USER AND CHANNEL OBJECT'S CREATION ***********************  
 **************************************************************************** 
 ************************************************************************** */

// Promise retrieved after user's initialization and network's configuration 
var starting_promise = private_key_promise.then((private_key_filename) => {

  console.log("Updating path to user's private key");
  options.private_key_path = path.join(options.private_key_path, private_key_filename);

  console.log("Create a client and set the wallet location");
  client = new hfc();

  return hfc.newDefaultKeyValueStore({ path: options.wallet_path });

}).then((wallet) => {

  console.log("Set wallet path, and associate user ", options.user_id, " with application");
  client.setStateStore(wallet);

  // Creating user with pre-generated crypto-material
  return client.createUser({username: options.user_id,
                            mspid: options.mspid,
                            cryptoContent: {privateKey: options.private_key_path,
                                            signedCert: options.ecert_path}});

}).then((user) => {

  console.log("Check user is enrolled, and set a query URL in the network");
  if (user === undefined || user.isEnrolled() === false) {
    console.error("User not defined, or not enrolled - error");
  }

  // Creating channel object passing the channel id to the constructor
  channel = client.newChannel(options.channel_id);
  var peerObj = client.newPeer(options.peer_url);

  // Adding a peer and an orderer to the channel to enable querying
  channel.addPeer(peerObj);
  channel.addOrderer(client.newOrderer(options.orderer_url));
  
  targets.push(peerObj);

  return;

}).catch((error) => {

  console.error(error);

});

/****************************************************************************
 ****************************************************************************  
 **************************** REST METHODS **********************************  
 **************************************************************************** 
 ************************************************************************** */

/*
 * @function : addContract
 * @action : invoke the addContract function on the blockchain which
 * add a new contract to the ledger
 * */
app.post('/addContract', function (req, res) {  

  starting_promise.then(function () {
    
    tx_id = client.newTransactionID();
    console.log("Assigning transaction_id: ", tx_id._transaction_id);

    var request = {
      targets: targets,
      chaincodeId: options.chaincode_id,
      fcn: 'addContract',
      args: [req.body.key.toString(), JSON.stringify(req.body.value)],
      chainId: options.channel_id,
      txId: tx_id
    };

    return channel.sendTransactionProposal(request);

  })
  .then(function (results) {

    var proposalResponses = results[0];
    var proposal = results[1];
    var header = results[2];

    let isProposalGood = false;
    
    // Check if transaction proposal was validated or not
    if (proposalResponses && proposalResponses[0].response &&
        proposalResponses[0].response.status === 200) {

      isProposalGood = true;
      console.log('transaction proposal was good');

    } else {
      console.error('transaction proposal was bad');
    }
    
    if (isProposalGood) {
      
      console.log(util.format(
        'Successfully sent Proposal and received ProposalResponse: Status - %s, message - "%s", metadata - "%s", endorsement signature: %s',
        proposalResponses[0].response.status, proposalResponses[0].response.message,
        proposalResponses[0].response.payload, proposalResponses[0].endorsement.signature));
      
      var request = {
        proposalResponses: proposalResponses,
        proposal: proposal,
        header: header
      };

      // set the transaction listener and set a timeout of 30sec
      // if the transaction did not get committed within the timeout period,
      // fail the test
      var transactionID = tx_id.getTransactionID();
      var eventPromises = [];

      let eh = client.newEventHub();
      eh.setPeerAddr(options.event_url);
      eh.connect();

      let txPromise = new Promise(function (resolve, reject) {

                                    let handle = setTimeout(function () {
                                                   eh.disconnect();
                                                   reject();
                                                 }, 30000);
                                    eh.registerTxEvent(transactionID, function (tx, code) {

                                      clearTimeout(handle);
                                      eh.unregisterTxEvent(transactionID);
                                      eh.disconnect();

                                      if (code !== 'VALID') {
                                        console.error(
                                          'The transaction was invalid, code = ' + code);
                                        reject();
                                      } else {
                                        console.log(
                                          'The transaction has been committed on peer ' +
                                            eh._ep._endpoint.addr);
                                        resolve();
                                      }
                                    });

                                  });
      
      eventPromises.push(txPromise);
      
      var sendPromise = channel.sendTransaction(request);

      return Promise.all([sendPromise].concat(eventPromises)).then(function (results) {
               console.log(' event promise all complete and testing complete');
               
               // the first returned value is from the 'sendPromise' which is from the 'sendTransaction()' call
               return proposalResponses[0].response; 
               
             }).catch(function (err) {
                        console.error('Failed to send transaction and get notifications within the timeout period.');
                        return 'Failed to send transaction and get notifications within the timeout period.';

                      });

    } else {
      
      console.error('Failed to send Proposal or receive valid response. Response null or status is not 200. exiting...');
      return util.format('%s', proposalResponses[0]);

    }

  }, function (err) {
       
       console.error('Failed to send proposal due to error: ' + err.stack ? err.stack : err);
       return 'Failed to send proposal due to error: ' + err.stack ? err.stack : err;

     })
  .then(function (response) {

    console.log(response);
    if (response.status === 200) {

      // Transaction is successfull, send an success message with payload as response to the http query
      res.send(JSON.stringify({success: util.format('%s', response.message), payload: response.payload}));

      console.log('Successfully sent transaction to the orderer.');
      return tx_id.getTransactionID();
      
    } else if (response.status != undefined) {

      // Transaction failed at ordering step, send an error message as response to the http query
      res.send(JSON.stringify({error: util.format('%s', response.message)}));

      console.error('Failed to order the transaction. Error code: ' + response.status);
      return 'Failed to order the transaction. Error code: ' + response.status;

    } else {

      // Filtering the useful message from the complete response
      var splittedStringArray = response.split("message: ");
      var usefulMessage = splittedStringArray[1].substring(0, splittedStringArray[1].length - 1);

      // Transaction failed at proposal step, send an error message as response to the http query
      res.send(JSON.stringify({error: util.format('%s', usefulMessage)}));
      return '';
      
    }

  }, function (err) {
       console.error('Failed to send transaction due to error: ' + err.stack ? err.stack : err);
       return 'Failed to send transaction due to error: ' + err.stack ? err.stack : err;
     });

});

/****************************************************************************
 ****************************************************************************  
 **************************** STARTING THE SERVER ***************************  
 **************************************************************************** 
 ************************************************************************** */

var server = app.listen(8081, function () {

               var host = server.address().address;
               var port = server.address().port;
               
               console.log("Example app listening at http://%s:%s", host, port);
               
             });
