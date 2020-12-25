/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Gateway, Wallets } = require('fabric-network');
const path = require('path');
const { buildCCPOrg1, buildCCPOrg2, buildWallet, prettyJSONString} = require('../../../test-application/javascript/AppUtil.js');

const myChannel = 'mychannel';
const myChaincodeName = 'auction';

async function bid(ccp,wallet,user,orgMSP,item,quantity,price) {
    try {

        const gateway = new Gateway();
      //connect using Discovery enabled

      await gateway.connect(ccp,
          { wallet: wallet, identity: user, discovery: { enabled: true, asLocalhost: true } });

        const network = await gateway.getNetwork(myChannel);
        const contract = network.getContract(myChaincodeName);

        console.log('\n--> Evaluate Transaction: get your client ID');
        let buyer = await contract.evaluateTransaction('GetSubmittingClientIdentity');
        console.log('*** Result:  Buyer ID is ' + buyer.toString());

        let bidData = { objectType: 'bid', quantity: parseInt(quantity) , price: parseInt(price), org: orgMSP, buyer: buyer.toString()};

        let bidTransaction = contract.createTransaction('Bid');
        bidTransaction.setEndorsingOrganizations(orgMSP);
        let tmapData = Buffer.from(JSON.stringify(bidData));
        bidTransaction.setTransient({
              bid: tmapData
            });

        let bidID = bidTransaction.getTransactionId();

        console.log('\n--> Submit Transaction: Create the bid that is stored in your private data collection of your organization');
        await bidTransaction.submit(item);
        console.log('*** Result: committed');
        console.log('*** Result ***SAVE THIS VALUE*** BidID: ' + bidID.toString());

        console.log('\n--> Evaluate Transaction: read the bid that was just created');
        let result = await contract.evaluateTransaction('QueryBid',item, bidID);
        console.log('*** Result:  Bid: ' + prettyJSONString(result.toString()));

        let transaction = contract.createTransaction('NewPublicBid');
        console.log('\n--> Submit Transaction: Add bid to the public book')
        await transaction.submit(item, bidID.toString());

        gateway.disconnect();
    } catch (error) {
        console.error(`******** FAILED to submit bid: ${error}`);
        if (error.stack) {
			console.error(error.stack);
		}
		process.exit(1);
	}
}

async function main() {
    try {

        if (process.argv[2] == undefined || process.argv[3] == undefined
            || process.argv[4] == undefined || process.argv[5] == undefined
            || process.argv[6] == undefined) {
            console.log("Usage: node bid.js org userID item quantity price");
            process.exit(1);
        }

        const org = process.argv[2];
        const user = process.argv[3];
        const item = process.argv[4];
        const quantity = process.argv[5];
        const price = process.argv[6];

        if (org == 'Org1' || org == 'org1') {

            const orgMSP = 'Org1MSP';
            const ccp = buildCCPOrg1();
            const walletPath = path.join(__dirname, 'wallet/org1');
            const wallet = await buildWallet(Wallets, walletPath);
            await bid(ccp,wallet,user,orgMSP,item,quantity,price);
        }
        else if (org == 'Org2' || org == 'org2') {

            const orgMSP = 'Org2MSP';
            const ccp = buildCCPOrg2();
            const walletPath = path.join(__dirname, 'wallet/org2');
            const wallet = await buildWallet(Wallets, walletPath);
            await bid(ccp,wallet,user,orgMSP,item,quantity,price);
        }  else {
            console.log("Usage: node bid.js org userID item quantity price");
            console.log("Org must be Org1 or Org2");
          }
    } catch (error) {
        console.error(`******** FAILED to run the application: ${error}`);
        process.exit(1);
    }
}


main();
