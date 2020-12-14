/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

/**
 * Application that shows events when creating and updating an asset
 *   -- How to register a contract listener for chaincode events
 *   -- How to get the chaincode event name and value from the chaincode event
 *   -- How to retrieve the transaction and block information from the chaincode event
 *   -- How to register a block listener for full block events
 *   -- How to retrieve the transaction and block information from the full block event
 *   -- How to register to recieve private data associated with transactions when
 *      registering a block listener
 *   -- How to retreive the private data from the full block event
 *   -- The listener will be notified of an event at anytime. Notice that events will
 *      be posted by the listener after the application activity causing the ledger change
 *      and during other application activity unrelated to the event
 *   -- How to connect to a Gateway that will not use events when submitting transactions.
 *      This may be useful when the application does not want to wait for the peer to commit
 *      blocks and notify the application.
 *
 * To see the SDK workings, try setting the logging to be displayed on the console
 * before executing this application.
 *        export HFC_LOGGING='{"debug":"console"}'
 * See the following on how the SDK is working with the Peer's Event Services
 * https://hyperledger-fabric.readthedocs.io/en/latest/peer_event_services.html
 *
 * See the following for more details on using the Node SDK
 * https://hyperledger.github.io/fabric-sdk-node/release-2.2/module-fabric-network.html
 */

// pre-requisites:
// - fabric-sample two organization test-network setup with two peers, ordering service,
//   and 2 certificate authorities
//         ===> from directory test-network
//         ./network.sh up createChannel -ca
//
// - Use the asset-transfer-events/chaincode-javascript chaincode deployed on
//   the channel "mychannel". The following deploy command will package, install,
//   approve, and commit the javascript chaincode, all the actions it takes
//   to deploy a chaincode to a channel.
//         ===> from directory test-network
//         ./network.sh deployCC -ccn events -ccp ../asset-transfer-events/chaincode-javacript/ -ccl javascript -ccep "OR('Org1MSP.peer','Org2MSP.peer')"
//
// - Be sure that node.js is installed
//         ===> from directory asset-transfer-sbe/application-javascript
//         node -v
// - npm installed code dependencies
//         ===> from directory asset-transfer-sbe/application-javascript
//         npm install
// - to run this test application
//         ===> from directory asset-transfer-sbe/application-javascript
//         node app.js

// NOTE: If you see an error like these:
/*

   Error in setup: Error: DiscoveryService: mychannel error: access denied

   OR

   Failed to register user : Error: fabric-ca request register failed with errors [[ { code: 20, message: 'Authentication failure' } ]]

	*/
// Delete the /fabric-samples/asset-transfer-sbe/application-javascript/wallet directory
// and retry this application.
//
// The certificate authority must have been restarted and the saved certificates for the
// admin and application user are not valid. Deleting the wallet store will force these to be reset
// with the new certificate authority.
//

// use this to set logging, must be set before the require('fabric-network');
process.env.HFC_LOGGING = '{"debug": "./debug.log"}';

const { Gateway, Wallets } = require('fabric-network');
const EventStrategies = require('fabric-network/lib/impl/event/defaulteventhandlerstrategies');
const FabricCAServices = require('fabric-ca-client');
const path = require('path');
const { buildCAClient, registerAndEnrollUser, enrollAdmin } = require('../../test-application/javascript/CAUtil.js');
const { buildCCPOrg1, buildWallet } = require('../../test-application/javascript/AppUtil.js');

const channelName = 'mychannel';
const chaincodeName = 'asset-transfer-events-javascript';

const org1 = 'Org1MSP';
const Org1UserId = 'appUser1';

const RED = '\x1b[31m\n';
const GREEN = '\x1b[32m\n';
const BLUE = '\x1b[34m';
const RESET = '\x1b[0m';

/**
 * Perform a sleep -- asynchronous wait
 * @param ms the time in milliseconds to sleep for
 */
function sleep(ms) {
	return new Promise((resolve) => setTimeout(resolve, ms));
}

async function initGatewayForOrg1(useCommitEvents) {
	console.log(`${GREEN}--> Fabric client user & Gateway init: Using Org1 identity to Org1 Peer${RESET}`);
	// build an in memory object with the network configuration (also known as a connection profile)
	const ccpOrg1 = buildCCPOrg1();

	// build an instance of the fabric ca services client based on
	// the information in the network configuration
	const caOrg1Client = buildCAClient(FabricCAServices, ccpOrg1, 'ca.org1.example.com');

	// setup the wallet to cache the credentials of the application user, on the app server locally
	const walletPathOrg1 = path.join(__dirname, 'wallet', 'org1');
	const walletOrg1 = await buildWallet(Wallets, walletPathOrg1);

	// in a real application this would be done on an administrative flow, and only once
	// stores admin identity in local wallet, if needed
	await enrollAdmin(caOrg1Client, walletOrg1, org1);
	// register & enroll application user with CA, which is used as client identify to make chaincode calls
	// and stores app user identity in local wallet
	// In a real application this would be done only when a new user was required to be added
	// and would be part of an administrative flow
	await registerAndEnrollUser(caOrg1Client, walletOrg1, org1, Org1UserId, 'org1.department1');

	try {
		// Create a new gateway for connecting to Org's peer node.
		const gatewayOrg1 = new Gateway();

		if (useCommitEvents) {
			await gatewayOrg1.connect(ccpOrg1, {
				wallet: walletOrg1,
				identity: Org1UserId,
				discovery: { enabled: true, asLocalhost: true }
			});
		} else {
			await gatewayOrg1.connect(ccpOrg1, {
				wallet: walletOrg1,
				identity: Org1UserId,
				discovery: { enabled: true, asLocalhost: true },
				eventHandlerOptions: EventStrategies.NONE
			});
		}


		return gatewayOrg1;
	} catch (error) {
		console.error(`Error in connecting to gateway for Org1: ${error}`);
		process.exit(1);
	}
}

async function main() {
	console.log(`${BLUE} **** START ****${RESET}`);
	try {
		let randomNumber = Math.floor(Math.random() * 1000) + 1;
		// use a random key so that we can run multiple times
		let assetKey = `item-${randomNumber}`;

		/** ******* Fabric client init: Using Org1 identity to Org1 Peer ******* */
		const gateway1Org1 = await initGatewayForOrg1(true); // transaction handling uses commit events
		const gateway2Org1 = await initGatewayForOrg1();

		try {
			//
			//  - - - - - -  C H A I N C O D E  E V E N T S
			//
			console.log(`${BLUE} **** CHAINCODE EVENTS ****${RESET}`);
			let transaction;
			let queryListener;
      let mainListener;
			const network1Org1 = await gateway1Org1.getNetwork(channelName);
			const contract1Org1 = network1Org1.getContract(chaincodeName);

			try {
				// first create a listener to be notified of chaincode code events
				// coming from the chaincode ID "events"
				queryListener = async (event) => {

					const asset = JSON.parse(event.payload.toString());
					// show the information available with the event
					// notice how we have access to the transaction information that produced this chaincode event

          try {
            const resultBuffer = await contract1Org1.evaluateTransaction('ReadAsset', asset.ID);
            let result = JSON.parse(resultBuffer.toString('utf8'));
            console.log(` Query results: ${resultBuffer}`);
          } catch (readError) {
            console.log(`${RED}<-- Failed: ReadAsset - ${readError}${RESET}`);
          }

					// notice how we have access to the full block that contains this transaction
					const eventBlock = eventTransaction.getBlockEvent();
					console.log(`*** block: ${eventBlock.blockNumber.toString()}`);
				};
				// now start the client side event service and register the listener
				console.log(`${GREEN}--> Start contract event stream to peer in Org1${RESET}`);
				await contract1Org1.addContractListener(queryListener);
			} catch (eventError) {
				console.log(`${RED}<-- Failed: Setup contract events - ${eventError}${RESET}`);
			}

      try {
        // first create a listener to be notified of chaincode code events
        // coming from the chaincode ID "events"
        mainListener = async (event) => {
          const asset = JSON.parse(event.payload.toString());
          // show the information available with the event
          console.log(`*** Event: ${event.eventName}:${asset.ID}`);
          // notice how we have access to the transaction information that produced this chaincode event

          switch (event.eventName) {
              case `CreateAsset`:
           	  try {
                // U P D A T E
                console.log(`${GREEN}--> Submit Transaction: UpdateAsset ${assetKey} update appraised value to 200`);
                transaction = contract1Org1.createTransaction('UpdateAsset');
                await transaction.submit(assetKey, 'blue', '10', 'Sam', '200');
                console.log(`${GREEN}<-- Submit UpdateAsset Result: committed, asset ${assetKey}${RESET}`);
              } catch (updateError) {
                console.log(`${RED}<-- Failed: UpdateAsset - ${updateError}${RESET}`);
              }
                  break;
              case `UpdateAsset`:
              try {
                // T R A N S F E R
                console.log(`${GREEN}--> Submit Transaction: TransferAsset ${assetKey} to Mary`);
                transaction = contract1Org1.createTransaction('TransferAsset');
                await transaction.submit(assetKey, 'Mary');
                console.log(`${GREEN}<-- Submit TransferAsset Result: committed, asset ${assetKey}${RESET}`);
              } catch (transferError) {
                console.log(`${RED}<-- Failed: TransferAsset - ${transferError}${RESET}`);
              }
                  break;
              case `TransferAsset`:
              try {
                // D E L E T E
                console.log(`${GREEN}--> Submit Transaction: DeleteAsset ${assetKey}`);
                transaction = contract1Org1.createTransaction('DeleteAsset');
                await transaction.submit(assetKey);
                console.log(`${GREEN}<-- Submit DeleteAsset Result: committed, asset ${assetKey}${RESET}`);
              } catch (deleteError) {
                console.log(`${RED}<-- Failed: DeleteAsset - ${deleteError}${RESET}`);
                if (deleteError.toString().includes('ENDORSEMENT_POLICY_FAILURE')) {
                  console.log(`${RED}Be sure that chaincode was deployed with the endorsement policy "OR('Org1MSP.peer','Org2MSP.peer')"${RESET}`)
                }
              }
                  break;
          }

        };
        // now start the client side event service and register the listener
        console.log(`${GREEN}--> Start contract event stream to peer in Org1${RESET}`);
        await contract1Org1.addContractListener(mainListener);
      } catch (eventError) {
        console.log(`${RED}<-- Failed: Setup contract events - ${eventError}${RESET}`);
      }

			try {
				// C R E A T E
				console.log(`${GREEN}--> Submit Transaction: CreateAsset, ${assetKey} owned by Sam${RESET}`);
				transaction = contract1Org1.createTransaction('CreateAsset');
				await transaction.submit(assetKey, 'blue', '10', 'Sam', '100');
				console.log(`${GREEN}<-- Submit CreateAsset Result: committed, asset ${assetKey}${RESET}`);
			} catch (createError) {
				console.log(`${RED}<-- Submit Failed: CreateAsset - ${createError}${RESET}`);
			}

      await sleep(5000);
			// all done with this listener
			contract1Org1.removeContractListener(queryListener);
      contract1Org1.removeContractListener(mainListener);

		} catch (runError) {
			console.error(`Error in transaction: ${runError}`);
			if (runError.stack) {
				console.error(runError.stack);
			}
		}
	} catch (error) {
		console.error(`Error in setup: ${error}`);
		if (error.stack) {
			console.error(error.stack);
		}
		process.exit(1);
	}

  await sleep(5000);
	console.log(`${BLUE} **** END ****${RESET}`);
	process.exit(0);
}
main();
