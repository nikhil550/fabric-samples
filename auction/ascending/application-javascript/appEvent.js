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

const channelName = 'mychannel';
const chaincodeName = 'auction';
const item = 'good9';
const auctionID = 'auction9';


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

              var auction = [];
              let AuctionID = event.payload.toString();
              var newAuction = setTimeout(async function auctionLoop(auction) {

                try {
                  let auctionResult = await contract.evaluateTransaction('QueryAuction', AuctionID);
                  let auctionJSON = JSON.parse(auctionResult);
                  var item = auctionJSON[0].item;
                  // update the current auction
                  for (let round = 0; round < auctionJSON.length; ++round) {

                    if (auction[round] == undefined) {
                      let auctionRound = new AuctionRound(JSON.stringify(auctionJSON[round].id),
                        JSON.stringify(auctionJSON[round].round),
                        JSON.stringify(auctionJSON[round].price),
                        JSON.stringify(auctionJSON[round].item),
                        JSON.stringify(auctionJSON[round].demand),
                        JSON.stringify(auctionJSON[round].quantity),
                        JSON.stringify(auctionJSON[round].sold));
                      auction[round] = auctionRound;
                      console.log(`*** New auction round: ${round}`);
                      console.log(auction[round]);

                      if (round > 0) {
                        console.log(`*** Previous round: ${round}`);
                        console.log(auction[round - 1]);
                      }
                    } else {
                      auction[round].updateAuction(JSON.stringify(auctionJSON[round].demand), JSON.stringify(auctionJSON[round].quantity), JSON.stringify(auctionJSON[round].sold));
                    };
                  };

                  // add bids and asks if the auction has not yet beed joined
                  for (let round = 0; round < auction.length; ++round) {
                    if (auction[round].joined == false) {

                      // query all bids on the item from your org
                      let result = await contract.evaluateTransaction('QueryBids', item);
                      let bids = JSON.parse(result);
                      for (let i = 0; i < bids.length; ++i) {
                        if (parseInt(auction[round].price) <= parseInt(bids[i].bid.price)) {
                          console.log(`*** Submitting bid for ${bids[i].bid.quantity} ${item} for round ${round}`);
                          // submit the bid auction
                          try {
                            let newBid = contract.createTransaction('SubmitBid');
                            await newBid.submit(AuctionID, round, bids[i].bid.quantity, bids[i].id);
                          } catch (error) {
                            console.log(`<-- Failed to submit bid: ${error}`);
                          };
                        };
                      };

                      // query asks on the item from your org
                      result = await contract.evaluateTransaction('QueryAsks', item);
                      let asks = JSON.parse(result);
                      for (let i = 0; i < asks.length; ++i) {
                        if ((parseInt(auction[round].price) >= parseInt(asks[i].ask.price)) && (parseInt(auction[round-1].price) < parseInt(asks[i].ask.price))) {
                          console.log(`*** Submitting ask for ${asks[i].ask.quantity} ${item} for round ${round}`);
                          // submit the ask auction
                          try {
                            let newAsk = contract.createTransaction('SubmitAsk');
                            await newAsk.submit(AuctionID, round, asks[i].ask.quantity, asks[i].id);
                          } catch (error) {
                            console.log(`<-- Failed to subit ask: ${error}`);
                          };
                        };
                      };
                      auction[round].join();
                    };
                  };

                  // query auction again after submitting bids
                  auctionResult = await contract.evaluateTransaction('QueryAuction', AuctionID);
                  auctionJSON = JSON.parse(auctionResult);
                  // update the current auction
                  for (let round = 0; round < auctionJSON.length; ++round) {

                    if (auction[round] == undefined) {
                      let auctionRound = new AuctionRound(JSON.stringify(auctionJSON[round].id),
                        JSON.stringify(auctionJSON[round].round),
                        JSON.stringify(auctionJSON[round].price),
                        JSON.stringify(auctionJSON[round].item),
                        JSON.stringify(auctionJSON[round].demand),
                        JSON.stringify(auctionJSON[round].quantity),
                        JSON.stringify(auctionJSON[round].sold));
                      auction[round] = auctionRound;
                    } else {
                      auction[round].updateAuction(JSON.stringify(auctionJSON[round].demand), JSON.stringify(auctionJSON[round].quantity), JSON.stringify(auctionJSON[round].sold));
                    };
                  };
                  // see if demand is greater than supply for the final round.
                  // if so, create a new round
                  let finalRound = auction.length - 1;
                  if (parseInt(auction[finalRound].demand) > parseInt(auction[finalRound].quantity)) {

                    try {
                      let transaction = contract.createTransaction('CreateNewRound');
                      let newRound = auction.length;
                      await transaction.submit(AuctionID, newRound);
                    } catch (error) {
                      console.log(`<-- Failed to create new round: ${error}`);
                    };
                  };

                  // go through rounds and try to close if supply
                  // is greater than demand
                  for (let round = 0; round < auction.length; ++round) {
                    if (parseInt(auction[round].demand) <= parseInt(auction[round].quantity) && (parseInt(auction[round].quantity) != 0)) {

                      // try to close the auction
                      try {
                        let closeRound = contract.createTransaction('CloseAuctionRound');
                        await closeRound.submit(AuctionID, round);
                      } catch (error) {
                        console.log(`<-- Failed to close round: ${error}`);
                      };
                    };
                  };
                  setTimeout(() => { auctionLoop(auction) }, 5000, auction);
                } catch (error) {
                 console.log(`<-- auction loop failed: ${error}`);
                }
              }, 5000, auction)

            } catch (error) {
              console.log(`<-- CreateAction event response failed: ${error}`);
            }
            break;

          case `CloseRound`:

            console.log(`<-- Close round event`)
            try {
              clearTimeout(newAuction);
              let AuctionID = event.payload.toString();
              let auctionResult = await contract.evaluateTransaction('QueryAuction', AuctionID);
              console.log('*** Full Closed Auction: ' + prettyJSONString(auctionResult.toString()));

              let endAuction = contract.createTransaction('EndAuction');
              await endAuction.submit(AuctionID);
            } catch (error) {
              console.log(`<-- Failed to close round: ${error}`);
            };

            break;

          case `EndAuction`:

            let AuctionID = event.payload.toString();
            let result = await contract.evaluateTransaction('QueryAuction', AuctionID);
            console.log('*** Result: Final auction round: ' + prettyJSONString(result.toString()));
            contract.removeContractListener(auctionListener);
            gatewayAuctionAdmin.disconnect()

            break;

        }
      };

      await contract.addContractListener(auctionListener);
      console.log(`<-- added contract listener`);


    } catch (eventError) {
      console.log(`<-- Failed: Setup event - ${eventError}`);
    }

    let transaction = contract.createTransaction('CreateAuction');
    await transaction.submit(auctionID, item, '20');

  } catch (error) {
    console.error(`******** FAILED to run the application: ${error}`);
    if (error.stack) {
      console.error(error.stack);
    }
    //    process.exit(1);
  }
}

main();
