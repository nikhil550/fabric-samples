/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { prettyJSONString } = require('../../../test-application/javascript/AppUtil.js');
const AuctionRound = require('./lib/auctionRound.js');
const { initGatewayForOrg2, sleep } = require('./lib/connect.js');

const channelName = 'mychannel';
const chaincodeName = 'auction';

const item = 'tickets';
const auctionID = 'auction3';

const mvccText = /MVCC_READ_CONFLICT/

async function main() {
  try {

    const adminGateway = await initGatewayForOrg2("auctionAdmin");
    const network = await adminGateway.getNetwork(channelName);
    const contract = network.getContract(chaincodeName);

    try {

      let auctionListener;
      auctionListener = async (event) => {

        switch (event.eventName) {
          case `CreateAuction`:
            try {

              var auction = [];
              var AuctionID = event.payload.toString();
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
                        JSON.stringify(auctionJSON[round].quantity));
                      auction[round] = auctionRound;
                      console.log(`*** New auction round: ${round}`);
                      console.log(auction[round]);

                      if (round > 0) {
                        console.log(`*** Previous round: ${round}`);
                        console.log(auction[round-1]);
                      }
                    } else {
                      auction[round].updateAuction(JSON.stringify(auctionJSON[round].demand), JSON.stringify(auctionJSON[round].quantity));
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

                          await sleep(Math.floor(Math.random() * 5000)+1000));
                          console.log(`*** Submitting bid for ${bids[i].bid.quantity} ${item} for round ${round}`);

                          // submit the bid auction
                          setTimeout(async function bid() {
                            try {
                              let newBid = contract.createTransaction('SubmitBid');
                              await newBid.submit(AuctionID, round, bids[i].bid.quantity, bids[i].id);
                            } catch (error) {
                              if (error.toString().match(mvccText) != null) {
                                console.log(error.name);
                                setTimeout(() => { bid() }, 5000);
                              } else {
                                console.log(`<-- Failed to submit bid: ${error}`);
                              };
                            };
                          }, 5000)
                        };
                      };

                      // query asks on the item from your org
                      result = await contract.evaluateTransaction('QueryAsks', item);
                      let asks = JSON.parse(result);
                      for (let i = 0; i < asks.length; ++i) {
                        if (parseInt(auction[round].price) >= parseInt(asks[i].ask.price) && (parseInt(auction[round-1].price) < parseInt(asks[i].ask.price))) {

                          await sleep(Math.floor(Math.random() * 7000));
                          console.log(`*** Submitting ask for ${asks[i].ask.quantity} ${item} for round ${round}`);

                          // submit the ask auction
                          setTimeout(async function ask() {
                            try {
                              let newAsk = contract.createTransaction('SubmitAsk');
                              await newAsk.submit(AuctionID, round, asks[i].ask.quantity, asks[i].id);
                            } catch (error) {
                              if (error.toString().match(mvccText) != null) {
                                console.log(error.name);
                                setTimeout(() => { ask() }, 5000);
                              } else {
                                console.log(`<-- Failed to subit ask: ${error}`);
                              };
                            };
                          }, 5000)
                        };
                      };
                      auction[round].join();
                    };
                  };

                  // see if demand is greater than supply for the final round.
                  // if so, create a new round
                  let finalRound = auction.length - 1;
                  if (parseInt(auction[finalRound].demand) > parseInt(auction[finalRound].quantity)) {

                    setTimeout(async function newRound() {
                      try {
                        let transaction = contract.createTransaction('CreateNewRound');
                        let newRound = auction.length;
                        await transaction.submit(AuctionID, newRound);
                      } catch (error) {
                        if (error.toString().match(mvccText) != null) {
                          console.log(error.name);
                          setTimeout(() => { newRound() }, 5000);
                        } else {
                          console.log(`<-- Failed to create new round: ${error}`);
                        }
                      };
                    }, 5000)
                  };

                  // go through rounds and try to close if supply
                  // is greater than demand
                  for (let round = 0; round < auction.length; ++round) {
                    if (parseInt(auction[round].demand) <= parseInt(auction[round].quantity) && (parseInt(auction[round].quantity) != 0)) {

                      // try to close the auction
                      setTimeout(async function closeRound() {
                        try {
                          let closeRound = contract.createTransaction('CloseAuctionRound');
                          await closeRound.submit(AuctionID, round);
                        } catch (error) {
                          if (error.toString().match(mvccText) != null) {
                            console.log(error.name);
                            setTimeout(() => { closeRound() }, 5000);
                          } else {
                            console.log(`<-- Failed to close round: ${error}`);
                          };
                        };
                      }, 5000)
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
            clearTimeout(newAuction);
            AuctionID = event.payload.toString();
            let auctionResult = await contract.evaluateTransaction('QueryAuction', AuctionID);
            console.log('*** Full Closed Auction: ' + prettyJSONString(auctionResult.toString()));

            setTimeout(async function endAuction() {
              try {
                let endAuction = contract.createTransaction('EndAuction');
                await endAuction.submit(AuctionID);
              } catch (error) {
                if (error.toString().match(mvccText) != null) {
                  console.log(error.name);
                  setTimeout(() => { endAuction() }, 5000);
                } else {
                  console.log(`<-- Failed to close round: ${error}`);
                };

              };
            }, 5000)

            break;

          case `EndAuction`:

            AuctionID = event.payload.toString();
            let result = await contract.evaluateTransaction('QueryAuction', AuctionID);
            console.log('*** Result: Final auction round: ' + prettyJSONString(result.toString()));
            contract.removeContractListener(auctionListener);
            adminGateway.disconnect()

            break;

        }
      };

      await contract.addContractListener(auctionListener);
      console.log(`<-- added auction listener`);


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
  }
}

main();
