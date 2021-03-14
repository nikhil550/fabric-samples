/*
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';


const { prettyJSONString } = require('../../../../test-application/javascript/AppUtil.js');

const channelName = 'mychannel';
const chaincodeName = 'auction';

exports.ask = async (gateway, orgMSP, item, quantity, price) => {

    try {

        const network = await gateway.getNetwork(channelName);
        const contract = network.getContract(chaincodeName);

        let seller = await contract.evaluateTransaction('GetSubmittingClientIdentity');
        console.log('*** Result:  Seller ID is ' + seller.toString());

        let privateAsk = { objectType: 'ask', quantity: parseInt(quantity), price: parseInt(price), org: orgMSP, seller: seller.toString() };
        let publicAsk = { objectType: 'ask', quantity: parseInt(quantity), org: orgMSP, seller: seller.toString() };

        let askTransaction = contract.createTransaction('Ask');
        askTransaction.setEndorsingOrganizations(orgMSP);
        let privateAskData = Buffer.from(JSON.stringify(privateAsk));
        let publicAskData = Buffer.from(JSON.stringify(publicAsk));
        askTransaction.setTransient({
            privateAsk: privateAskData,
            publicAsk: publicAskData
        });

        let askID = askTransaction.getTransactionId();

        console.log('\n--> Submit Transaction: Create the ask that is stored in your private data collection of your organization');
        await askTransaction.submit(item);
        console.log('*** Result: committed');
        console.log('*** Result ***SAVE THIS VALUE*** AskID: ' + askID.toString());

        console.log('\n--> Evaluate Transaction: read the ask that was just created');
        let result = await contract.evaluateTransaction('QueryAsk', item, askID);
        console.log('*** Result:  Ask: ' + prettyJSONString(result.toString()));

        let transaction = contract.createTransaction('NewPublicAsk');
        console.log('\n--> Submit Transaction: Add bid to the public book')
        await transaction.submit(item, askID.toString());

        //    gateway.disconnect();
        return askID;
    } catch (error) {
        console.error(`******** FAILED to submit ask: ${error}`);
        if (error.stack) {
            console.error(error.stack);
        }
        process.exit(1);
    }
}

exports.bid = async (gateway, orgMSP, item, quantity, price) => {

    try {

        const network = await gateway.getNetwork(channelName);
        let contract = network.getContract(chaincodeName);

        let buyer = await contract.evaluateTransaction('GetSubmittingClientIdentity');
        console.log('*** Result:  Buyer ID is ' + buyer.toString());

        let privateBid = { objectType: 'bid', quantity: parseInt(quantity), price: parseInt(price), org: orgMSP, buyer: buyer.toString() };
        let publicBid = { objectType: 'bid', quantity: parseInt(quantity), org: orgMSP, buyer: buyer.toString() };


        let bidTransaction = contract.createTransaction('Bid');
        bidTransaction.setEndorsingOrganizations(orgMSP);
        let privateBidData = Buffer.from(JSON.stringify(privateBid));
        let publicBidData = Buffer.from(JSON.stringify(publicBid));
        bidTransaction.setTransient({
            privateBid: privateBidData,
            publicBid: publicBidData
        });

        let bidID = bidTransaction.getTransactionId();

        console.log('\n--> Submit Transaction: Create the bid that is stored in your private data collection of your organization');
        await bidTransaction.submit(item);
        console.log('*** Result: committed');
        console.log('*** Result ***SAVE THIS VALUE*** BidID: ' + bidID.toString());

        console.log('\n--> Evaluate Transaction: read the bid that was just created');
        let result = await contract.evaluateTransaction('QueryBid', item, bidID);
        console.log('*** Result:  Bid: ' + prettyJSONString(result.toString()));

        let transaction = contract.createTransaction('NewPublicBid');
        console.log('\n--> Submit Transaction: Add ask to the public book')
        await transaction.submit(item, bidID.toString());

        //      gateway.disconnect();
        return bidID;
    } catch (error) {
        console.error(`******** FAILED to submit bid: ${error}`);
        if (error.stack) {
            console.error(error.stack);
        }
        process.exit(1);
    }
}
