/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { prettyJSONString } = require('../../../test-application/javascript/AppUtil.js');
const AuctionRound = require('./lib/auctionRound.js');
const { initGatewayForOrg1, initGatewayForOrg2, sleep } = require('./lib/connect.js');
const { bid, ask } = require('./lib/bidAsk.js');
const { auctionBuyer, auctionSeller } = require('./lib/event.js');

const channelName = 'mychannel';
const chaincodeName = 'auction';
const item = 'good10';
const auctionID = 'auction10';


async function main() {
  try {

    const gatewayBidder1 = await initGatewayForOrg1("bidder1")
    await bid(gatewayBidder1, "Org1MSP", item, "20", "20");
    gatewayBidder1.disconnect();

    const gatewayBidder2 = await initGatewayForOrg1("bidder2")
    await bid(gatewayBidder2, "Org1MSP", item, "20", "40");
    gatewayBidder2.disconnect();

    const gatewayBidder3 = await initGatewayForOrg1("bidder3")
    await bid(gatewayBidder3, "Org1MSP", item, "20", "60");
    gatewayBidder3.disconnect();

    const gatewayBidder4 = await initGatewayForOrg1("bidder4")
    await bid(gatewayBidder4, "Org1MSP", item, "20", "80");
    gatewayBidder4.disconnect();

    const gatewayBidder5 = await initGatewayForOrg1("bidder5")
    await bid(gatewayBidder5, "Org1MSP", item, "20", "100");
    gatewayBidder5.disconnect();

    const gatewaySeller1 = await initGatewayForOrg1("seller1")
    await ask(gatewaySeller1, "Org1MSP", item, "20", "30");
    gatewaySeller1.disconnect();

    const gatewaySeller2 = await initGatewayForOrg1("seller2")
    await ask(gatewaySeller2, "Org1MSP", item, "20", "50");
    gatewaySeller2.disconnect();

    const gatewaySeller3 = await initGatewayForOrg1("seller3")
    await ask(gatewaySeller3, "Org1MSP", item, "20", "70");
    gatewaySeller3.disconnect();

    const gatewaySeller4 = await initGatewayForOrg1("seller4")
    await ask(gatewaySeller4, "Org1MSP", item, "20", "90");
    gatewaySeller4.disconnect();

    const gatewayAuctionAdmin = await initGatewayForOrg1("auctionAdmin");
    const network = await gatewayAuctionAdmin.getNetwork(channelName);
    const contract = network.getContract(chaincodeName);

    try {

      let auctionListener;
      auctionListener = async (event) => {

        switch (event.eventName) {
          case `CreateAuction`:

            try {

              var activeAuction = true;

              var auction = [];
              while (activeAuction) {
                await sleep(5000);
                let auctionResult = await contract.evaluateTransaction('QueryAuction', auctionID);
                console.log('*** Result: Auction: ' + prettyJSONString(auctionResult.toString()));
                let auctionJSON = JSON.parse(auctionResult);
                // update the current auction
                for (let round = 0; round < auctionJSON.length; ++round) {
                  if (auction[round] == undefined) {
                    let auctionRound = new AuctionRound(JSON.stringify(auctionJSON[round].id),
                      JSON.stringify(auctionJSON[round].round),
                      JSON.stringify(auctionJSON[round].price),
                      JSON.stringify(auctionJSON[round].item),
                      JSON.stringify(auctionJSON[round].demand),
                      JSON.stringify(auctionJSON[round].quantity));
                    auction[round] = auctionRound;
                  } else {
                    auction[round].updateAuction(JSON.stringify(auctionJSON[round].demand), JSON.stringify(auctionJSON[round].quantity));
                  };
                };
                console.log(auction);

                // add bids and asks if the auction has not yet beed joined
                for (let round = 0; round < auction.length; ++round) {
                  if (auction[round].joined == false) {

                    // query all bids on the item from your org
                    let result = await contract.evaluateTransaction('QueryBids', item);
                    //console.log('*** Result: Bid: ' + prettyJSONString(result.toString()));
                    let bids = JSON.parse(result);
                    for (let i = 0; i < bids.length; ++i) {
                      if (auction[round].price <= bids[i].bid.price) {
                        // submit the bid auction
                        try {
                          let newBid = contract.createTransaction('SubmitBid');
                          await newBid.submit(auctionID, round, bids[i].bid.quantity, bids[i].id);
                        } catch (error) {
                          console.log(`<-- Failed to submit bid: ${error}`);
                        };
                      };
                    };

                    // query asks on the item from your org
                    result = await contract.evaluateTransaction('QueryAsks', item);
                    let asks = JSON.parse(result);
                    for (let i = 0; i < asks.length; ++i) {
                      if (auction[round].price >= asks[i].ask.price) {
                        // submit the ask auction
                        try {
                          let newAsk = contract.createTransaction('SubmitAsk');
                          await newAsk.submit(auctionID, round, asks[i].ask.quantity, asks[i].id);
                        } catch (error) {
                          console.log(`<-- Failed to subit ask: ${error}`);
                        };
                      };
                    };
                    auction[round].join();
                  };
                };

                // see if demand is greater than supply for the final round.
                // if so, create a new round
                let finalRound = auction.length - 1;
                if (auction[finalRound].demand > auction[finalRound].quantity) {

                  try {
                    let transaction = contract.createTransaction('CreateNewRound');
                    let newRound = auction.length;
                    await transaction.submit(auctionID, newRound);
                  } catch (error) {
                    console.log(`<-- Failed to create new round: ${error}`);
                  };
                };

                // go through rounds and try to close if demand
                // is greater than supply
                for (let round = 0; round < auction.length; ++round) {
                  if (auction[round].demand <= auction[round].quantity && auction[round].quantity != 0) {

                    // try to close the auction
                    try {
                      let closeRound = contract.createTransaction('CloseAuctionRound');
                      await closeRound.submit(auctionID, round);
                    } catch (error) {
                      console.log(`<-- Failed to close round: ${error}`);
                    };
                  };
                };

                console.log(`auction loop:`);
              };

            } catch (error) {
              console.log(`<-- CreateAction event response failed: ${error}`);
            }
            break;

          case `CloseRound`:

            var activeAuction = false;
            console.log(`<-- Close round event`)
            try {
              let endAuction = contract.createTransaction('EndAuction');
              await endAuction.submit(auctionID);
            } catch (error) {
              console.log(`<-- Failed to close round: ${error}`);
            };

            break;

          case `EndAuction`:

            let result = await contract.evaluateTransaction('QueryAuction',auctionID);
            console.log('*** Result: Ended Auction: ' + prettyJSONString(result.toString()));
            
            break;

        }
      };

      await contract.addContractListener(auctionListener);
      console.log(`<-- added contract listener`);

      let transaction = contract.createTransaction('CreateAuction');
      await transaction.submit(auctionID, item, '20');  

    } catch (eventError) {
      console.log(`<-- Failed: Setup event - ${eventError}`);
    }



    // all done with this listener
    ///		contract.removeContractListener(auctionListener);
    //    gatewayBidder1.disconnect()
  } catch (error) {
    console.error(`******** FAILED to run the application: ${error}`);
    if (error.stack) {
      console.error(error.stack);
    }
    //    process.exit(1);
  }
}

main();
