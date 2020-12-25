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

async function closeAuction(ccp,wallet,user,auctionID,round) {
    try {

        const gateway = new Gateway();
      //connect using Discovery enabled

      await gateway.connect(ccp,
          { wallet: wallet, identity: user, discovery: { enabled: true, asLocalhost: true } });

        const network = await gateway.getNetwork(myChannel);
        const contract = network.getContract(myChaincodeName);

        let statefulTxn = contract.createTransaction('CloseAuctionRound');

        console.log('\n--> Submit Transaction: close auction round');
        await statefulTxn.submit(auctionID,round);
        console.log('*** Result: committed');

        console.log('\n--> Evaluate Transaction: query the updated auction');
        let result = await contract.evaluateTransaction('QueryAuctionRound',auctionID,round);
        console.log('*** Result: Auction: ' + prettyJSONString(result.toString()));

        gateway.disconnect();
    } catch (error) {
        console.error(`******** FAILED to close auction: ${error}`);
        process.exit(1);
	}
}

async function main() {
    try {

        if (process.argv[2] == undefined || process.argv[3] == undefined
            || process.argv[4] == undefined || process.argv[5] == undefined) {
            console.log("Usage: node closeAuction.js org userID auctionID, round");
            process.exit(1);
        }

        const org = process.argv[2]
        const user = process.argv[3];
        const auctionID = process.argv[4];
        const round = process.argv[5];

        if (org == 'Org1' || org == 'org1') {

            const orgMSP = 'Org1MSP';
            const ccp = buildCCPOrg1();
            const walletPath = path.join(__dirname, 'wallet/org1');
            const wallet = await buildWallet(Wallets, walletPath);
            await closeAuction(ccp,wallet,user,auctionID,round);
        }
        else if (org == 'Org2' || org == 'org2') {

            const orgMSP = 'Org2MSP';
            const ccp = buildCCPOrg2();
            const walletPath = path.join(__dirname, 'wallet/org2');
            const wallet = await buildWallet(Wallets, walletPath);
            await closeAuction(ccp,wallet,user,auctionID,round);
        }  else {
            console.log("Usage: node closeAuction.js org userID auctionID round");
            console.log("Org must be Org1 or Org2");
          }
    } catch (error) {
		console.error(`******** FAILED to run the application: ${error}`);
    if (error.stack) {
        console.error(error.stack);
    }
    process.exit(1);
    }
}


main();
