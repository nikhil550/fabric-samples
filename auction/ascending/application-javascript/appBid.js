/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Gateway, Wallets } = require('fabric-network');
const path = require('path');
const { buildCCPOrg1, buildCCPOrg2, buildWallet } = require('../../../test-application/javascript/AppUtil.js');

const myChannel = 'mychannel';
const myChaincodeName = 'auction';


function prettyJSONString(inputString) {
    if (inputString) {
        return JSON.stringify(JSON.parse(inputString), null, 2);
    }
    else {
        return inputString;
    }
}


async function initGatewayForOrg1(user) {
	console.log(`Fabric client user & Gateway init: Using Org1 identity to Org1 Peer`);

	const ccpOrg1 = buildCCPOrg1();

	const walletPathOrg1 = path.join(__dirname, 'wallet/org1');
	const walletOrg1 = await buildWallet(Wallets, walletPathOrg1);


	try {
		// Create a new gateway for connecting to Org's peer node.
		const gatewayOrg1 = new Gateway();
		//connect using Discovery enabled
		await gatewayOrg1.connect(ccpOrg1,
			{ wallet: walletOrg1, identity: user, discovery: { enabled: true, asLocalhost: true } });

		return gatewayOrg1;
	} catch (error) {
		console.error(`Error in connecting to gateway for Org1: ${error}`);
		process.exit(1);
	}
}

async function initGatewayForOrg2(user) {

	console.log(`--> Fabric client user & Gateway init: Using Org2 identity to Org2 Peer`);

  const ccpOrg2 = buildCCPOrg2();

	const walletPathOrg2 = path.join(__dirname, 'wallet/org2');
  const walletOrg2 = await buildWallet(Wallets, walletPathOrg2);

	try {
		// Create a new gateway for connecting to Org's peer node.
		const gatewayOrg2 = new Gateway();
		await gatewayOrg2.connect(ccpOrg2,
			{ wallet: walletOrg2, identity: user, discovery: { enabled: true, asLocalhost: true } });

		return gatewayOrg2;
	} catch (error) {
		console.error(`Error in connecting to gateway for Org2: ${error}`);
		process.exit(1);
	}
}


async function ask(gateway,orgMSP,item,quantity,price) {
    try {

        const network = await gateway.getNetwork(myChannel);
        const contract = network.getContract(myChaincodeName);

        console.log('\n--> Evaluate Transaction: get your client ID');
        let seller = await contract.evaluateTransaction('GetID');
        console.log('*** Result:  Seller ID is ' + seller.toString());

        let askData = { objectType: 'ask', quantity: parseInt(quantity) , price: parseInt(price), org: orgMSP, seller: seller.toString()};

        let statefulTxn = contract.createTransaction('Ask');
        statefulTxn.setEndorsingOrganizations(orgMSP);
        let tmapData = Buffer.from(JSON.stringify(askData));
        statefulTxn.setTransient({
              ask: tmapData
            });

        let askID = statefulTxn.getTransactionId();

        console.log('\n--> Submit Transaction: Create the ask that is stored in your private data collection of your organization');
        await statefulTxn.submit(item);
        console.log('*** Result: committed');
        console.log('*** Result ***SAVE THIS VALUE*** AskID: ' + askID.toString());

        console.log('\n--> Evaluate Transaction: read the ask that was just created');
        let result = await contract.evaluateTransaction('QueryAsk',item, askID);
        console.log('*** Result:  Ask: ' + prettyJSONString(result.toString()));

        gateway.disconnect();
    } catch (error) {
        console.error(`******** FAILED to submit ask: ${error}`);
        if (error.stack) {
			console.error(error.stack);
		}
		process.exit(1);
	}
}


async function bid(gateway,orgMSP,item,quantity,price) {
    try {

        const network = await gateway.getNetwork(myChannel);
        const contract = network.getContract(myChaincodeName);

        console.log('\n--> Evaluate Transaction: get your client ID');
        let buyer = await contract.evaluateTransaction('GetID');
        console.log('*** Result:  Buyer ID is ' + buyer.toString());

        let bidData = { objectType: 'bid', quantity: parseInt(quantity) , price: parseInt(price), org: orgMSP, buyer: buyer.toString()};

        let statefulTxn = contract.createTransaction('Bid');
        statefulTxn.setEndorsingOrganizations(orgMSP);
        let tmapData = Buffer.from(JSON.stringify(bidData));
        statefulTxn.setTransient({
              bid: tmapData
            });

        let bidID = statefulTxn.getTransactionId();

        console.log('\n--> Submit Transaction: Create the bid that is stored in your private data collection of your organization');
        await statefulTxn.submit(item);
        console.log('*** Result: committed');
        console.log('*** Result ***SAVE THIS VALUE*** BidID: ' + bidID.toString());

        console.log('\n--> Evaluate Transaction: read the bid that was just created');
        let result = await contract.evaluateTransaction('QueryBid',item, bidID);
        console.log('*** Result:  Bid: ' + prettyJSONString(result.toString()));

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

      const gatewayBidder1 = await initGatewayForOrg1("seller1")
      await bid(gatewayBidder1,"Org1MSP","tickets","20","20");
      gatewaySeller1.disconnect();

      const gatewayBidder2 = await initGatewayForOrg1("seller1")
      await bid(gatewayBidder2,"Org1MSP","tickets","20","40");
      gatewaySeller1.disconnect();

      const gatewayBidder3 = await initGatewayForOrg2("seller1")
      await bid(gatewayBidder3,"Org2MSP","tickets","20","60");
      gatewaySeller1.disconnect();

      const gatewayBidder4 = await initGatewayForOrg2("seller1")
      await bid(gatewayBidder4,"Org2MSP","tickets","20","80");
      gatewaySeller1.disconnect();

      const gatewayBidder5 = await initGatewayForOrg2("seller1")
      await bid(gatewayBidder5,"Org2MSP","tickets","20","100");
      gatewaySeller1.disconnect();

      const gatewaySeller1 = await initGatewayForOrg1("seller1")
      await bid(gatewaySeller1,"Org1MSP","tickets","20","30");
      gatewaySeller1.disconnect();

      const gatewaySeller2 = await initGatewayForOrg1("seller1")
      await bid(gatewaySeller2,"Org1MSP","tickets","20","50");
      gatewaySeller1.disconnect();

      const gatewaySeller3 = await initGatewayForOrg2("seller1")
      await bid(gatewaySeller3,"Org2MSP","tickets","20","70");
      gatewaySeller1.disconnect();

      const gatewaySeller4 = await initGatewayForOrg2("seller1")
      await bid(gatewaySeller4,"Org2MSP","tickets","20","90");
      gatewaySeller1.disconnect();

    } catch (error) {
		console.error(`******** FAILED to run the application: ${error}`);
    if (error.stack) {
        console.error(error.stack);
    }
    process.exit(1);
    }
}


main();
