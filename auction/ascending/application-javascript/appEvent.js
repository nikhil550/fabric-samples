/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { prettyJSONString} = require('../../../test-application/javascript/AppUtil.js');
const State = require('./lib/auctionRound.js');
const { initGatewayForOrg1, initGatewayForOrg2, sleep} = require('./lib/connect.js');
const { bid, ask } = require('./lib/bidAsk.js');
const { auctionBuyer, auctionSeller} = require('./lib/event.js');

const channelName = 'mychannel';
const chaincodeName = 'auction';
const item = 'good1';
const auctionID = 'auction1';


async function main() {
    try {

      const gatewayBidder1 = await initGatewayForOrg1("bidder1")
      const bidIDBidder1 = await bid(gatewayBidder1,"Org1MSP",item,"20","20");

      const gatewayBidder2 = await initGatewayForOrg1("bidder2")
      const bidIDBidder2 = await bid(gatewayBidder2,"Org1MSP",item,"20","40");

      const gatewayBidder3 = await initGatewayForOrg2("bidder3")
      const bidIDBidder3 = await bid(gatewayBidder3,"Org2MSP",item,"20","60");

      const gatewayBidder4 = await initGatewayForOrg2("bidder4")
      const bidIDBidder4 = await bid(gatewayBidder4,"Org2MSP",item,"20","80");

      const gatewayBidder5 = await initGatewayForOrg2("bidder5")
      const bidIDBidder5 = await bid(gatewayBidder5,"Org2MSP",item,"20","100");

      const gatewaySeller1 = await initGatewayForOrg1("seller1")
      const askIDSeller1 = await ask(gatewaySeller1,"Org1MSP",item,"20","30");

      const gatewaySeller2 = await initGatewayForOrg1("seller2")
      const askIDSeller2 = await ask(gatewaySeller2,"Org1MSP",item,"20","50");

      const gatewaySeller3 = await initGatewayForOrg2("seller3")
      const askIDSeller3 = await ask(gatewaySeller3,"Org2MSP",item,"20","70");

      const gatewaySeller4 = await initGatewayForOrg2("seller4")
      const askIDSeller4 = await ask(gatewaySeller4,"Org2MSP",item,"20","90");


			// setup event for bidder1
//			const gatewayBidder = await initGatewayForOrg1("bidder1")
			auctionBuyer(gatewayBidder1,bidIDBidder1,activeAuction);
		//	auctionBuyer(gatewayBidder2,bidIDBidder2);
		//	auctionBuyer(gatewayBidder3,bidIDBidder3);
		//	auctionBuyer(gatewayBidder4,bidIDBidder4);
		//	auctionBuyer(gatewayBidder5,bidIDBidder5);
		//	auctionSeller(gatewaySeller1,askIDSeller1);
		//	auctionSeller(gatewaySeller2,askIDSeller3);
		//	auctionSeller(gatewaySeller3,askIDSeller3);
		//	auctionSeller(gatewaySeller4,askIDSeller4);

			//	const gatewaySeller = await initGatewayForOrg1("seller1");
				const network = await gatewaySeller1.getNetwork(channelName);
				const contract = network.getContract(chaincodeName);

    		let transaction = contract.createTransaction('CreateAuction');
    		await transaction.submit(auctionID, item, '20');

        // C R E A T E


    //    await sleep(30000);

        var activeAuction = true;

        while (activeAuction) {
          await sleep(5000);
          let result = await contract.evaluateTransaction('QueryAuction',auctionID);
          console.log('*** Result: Auction: ' + prettyJSONString(result.toString()));
          let auction = JSON.parse(result.toString());
          var activeAuction = new Person(auction.id,auction.round, auction.item, auction.demand, auction.quantity);

          console.log(`auction loop:`);
            // console.log(`waiting`)
          };



    		// all done with this listener
    ///		contract.removeContractListener(auctionListener);
    //    gatewayBidder1.disconnect();

    } catch (error) {
		console.error(`******** FAILED to run the application: ${error}`);
    if (error.stack) {
        console.error(error.stack);
    }
//    process.exit(1);
    }
}

main();
