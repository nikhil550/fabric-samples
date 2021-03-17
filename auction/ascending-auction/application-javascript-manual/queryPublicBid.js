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

async function queryPublicBid(ccp,wallet,user,item,askSell,bidID) {
    try {

        const gateway = new Gateway();
      //connect using Discovery enabled

      await gateway.connect(ccp,
          { wallet: wallet, identity: user, discovery: { enabled: true, asLocalhost: true } });

        const network = await gateway.getNetwork(myChannel);
        const contract = network.getContract(myChaincodeName);

        console.log('\n--> Evaluate Transaction: read bid from private data store');
        let result = await contract.evaluateTransaction('QueryPublic',item,askSell,bidID);
        console.log(result.toString());

        gateway.disconnect();
    } catch (error) {
        console.error(`******** FAILED to submit bid: ${error}`);
	}
}

async function main() {
    try {

        if (process.argv[2] == undefined || process.argv[3] == undefined
            || process.argv[4] == undefined || process.argv[5] == undefined) {
            console.log("Usage: node bid.js org userID item bidID");
            process.exit(1);
        }

        const org = process.argv[2];
        const user = process.argv[3];
        const item = process.argv[4];
        const askSell = process.argv[5];
        const bidID = process.argv[6];

        if (org == 'Org1' || org == 'org1') {

            const orgMSP = 'Org1MSP';
            const ccp = buildCCPOrg1();
            const walletPath = path.join(__dirname, 'wallet/org1');
            const wallet = await buildWallet(Wallets, walletPath);
            await queryPublicBid(ccp,wallet,user,item,askSell,bidID);
        }
        else if (org == 'Org2' || org == 'org2') {

            const orgMSP = 'Org2MSP';
            const ccp = buildCCPOrg2();
            const walletPath = path.join(__dirname, 'wallet/org2');
            const wallet = await buildWallet(Wallets, walletPath);
            await queryPublicBid(ccp,wallet,user,item,askSell,bidID);
        } else {
            console.log("Usage: node bid.js org userID item bidID");
            console.log("Org must be Org1 or Org2");
          }
    } catch (error) {
        console.error(`******** FAILED to run the application: ${error}`);
    }
}


main();
