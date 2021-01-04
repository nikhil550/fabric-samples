/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const sleep = require('./connect.js');
const { prettyJSONString} = require('../../../../test-application/javascript/AppUtil.js');


const channelName = 'mychannel';
const chaincodeName = 'auction';

exports.auctionBuyer = async (gateway, bidID, activeAuction) => {

  const network = await gateway.getNetwork(channelName);
  const contract = network.getContract(chaincodeName);

  //let activeAuction = false;
  const regex = /cat jumps over/


  try {

  let auctionListener;
  auctionListener = async (event) => {

    let randomTime = Math.floor(Math.random() * 15000) + 1000
    await sleep(randomTime)

    switch (event.eventName) {
      case `CreateAuction`:

       try {
        // let activeAuction = true;

         let auction = JSON.parse(event.payload.toString());

         console.log(`New Auction: ${prettyJSONString(event.payload.toString())}`);
        //activeAuctionRound = new AuctionRound(auction.id,auction.round,auction.item,auction.demand,auction.quantity);
        //activeAuctionRound.item = auction.item;
        //activeAuctionRound.id = auction.id;
         activeAuction = true;
         console.log(`New Auction: ${activeAuction}`);
         // query bid
         let bid = await contract.evaluateTransaction('QueryBid',auction.item,bidID);
         let bidJSON = JSON.parse(bid.toString('utf8'));
         if (bidJSON.price >= auction.price) {
            let bidTransaction = contract.createTransaction('SubmitBid');
            await bidTransaction.submit(auction.id,auction.round,bidJSON.quantity,bidID);

            await sleep(5000);

            let transaction = contract.createTransaction('CloseAuctionRound');
            await transaction.submit(auction.id,auction.round);
          }

        } catch (error) {
        console.log(`<-- CreateAction event response failed: ${error}`);
      }
      break;

      case `CreateNewRound`:

        try {

          let auction = JSON.parse(event.payload.toString());
          console.log(`New Round: ${prettyJSONString(event.payload.toString())}`);
          console.log(`New round: ${activeAuction}`);
          // query bid
          let bid = await contract.evaluateTransaction('QueryBid',auction.item,bidID);
          let bidJSON = JSON.parse(bid.toString('utf8'));
          if (bidJSON.price >= auction.price) {
            let bidTransaction = contract.createTransaction('SubmitBid');
            await bidTransaction.submit(auction.id,auction.round,bidJSON.quantity,bidID);

            await sleep(5000);

            let transaction = contract.createTransaction('CloseAuctionRound');
            await transaction.submit(auction.id,auction.round);
           }

         } catch (error) {
         console.log(`<-- CreateNewRound event response failed: ${error}`);
       }
       break;

       case `EndAuction`:

         console.log(`Auction ended ${event.payload}`);
        break;

      }
    };

    await contract.addContractListener(auctionListener);
    console.log(`<-- added contract listener`);

  } catch (eventError) {
    console.log(`<-- Failed: Setup event - ${eventError}`);
  }
};



exports.auctionSeller = async (gateway, askID) => {

  const network = await gateway.getNetwork(channelName);
  const contract = network.getContract(chaincodeName);

 // let activeAuction = false;

function auctionRound(id, round, item, demand, quantity) {
    this.id = id;
    this.round = round;
    this.item = item;
    this.demand = demand;
    this.quantity = quantity;
  };

  let activeAuctionRound = {};

  try {

  let auctionListener;
  auctionListener = async (event) => {

    let randomTime = Math.floor(Math.random() * 25000) + 1000;
    await sleep(randomTime);

    switch (event.eventName) {

      case `CreateAuction`:

       try {

       // let activeAuction = true;
        let auction = JSON.parse(event.payload.toString());
        console.log(`New Auction: ${prettyJSONString(event.payload.toString())}`);
        //activeAuctionRound = new auctionRound(auction.id,auction.round,auction.item,auction.demand,auction.quantity);
        //console.log(`New Auction: ${activeAuctionRound}`);
        // query ask
        let ask = await contract.evaluateTransaction('QueryAsk',auction.item,askID);
        let askJSON = JSON.parse(ask.toString('utf8'));
        if (askJSON.price <= auction.price) {
            let askTransaction = contract.createTransaction('SubmitAsk');
            await askTransaction.submit(auction.id,auction.round,askJSON.quantity,askID);
          }

        } catch (error) {
        console.log(`<-- CreateAction event response failed: ${error.name}`);
      }
      break;

      case `CreateNewRound`:

        try {

         let auction = JSON.parse(event.payload.toString());
         console.log(`New Round: ${prettyJSONString(event.payload.toString())}`);
         // query ask
         let ask = await contract.evaluateTransaction('QueryAsk',auction.item,askID);
         let askJSON = JSON.parse(ask.toString('utf8'));
         if (askJSON.price <= auction.price) {
             let transaction = contract.createTransaction('SubmitAsk');
             await transaction.submit(auction.id,auction.round,askJSON.quantity,askID);
           }

         } catch (error) {
         console.log(`<-- CreateNewRound event response failed: ${error.name}`);
       }
       break;

      case `newBid`:

        try {

         let bidEvent = JSON.parse(event.payload.toString());
         console.log(`New Bid: ${prettyJSONString(event.payload.toString())}`);
         if (bidEvent.demand > bidEvent.quantity ) {
          let transaction = contract.createTransaction('CreateNewRound');
          let newRound = bidEvent.round + 1
          await transaction.submit(bidEvent.id,newRound);
          }

         } catch (error) {
         console.log(`<-- newBid event response failed: ${error.name}`);
       }
       break;

       case `CloseAuction`:

         try {

          let closedAuction = JSON.parse(event.payload.toString());
          console.log(`New Bid: ${prettyJSONString(event.payload.toString())}`);
          let transaction = contract.createTransaction('EndAuction');
          await transaction.submit(closedAuction.id);

          } catch (error) {
          console.log(`<-- CloseAuction event response failed: ${error.name}`);
        }

      case `EndAuction`:
    //  activeAuction = false
        console.log(`Auction ended ${event.payload}`);
       break;
      };
    };

    await contract.addContractListener(auctionListener);
    console.log(`<-- added contract listener`);


  } catch (eventError) {
    console.log(`<-- Failed: Setup event - ${eventError}`);
  }
};
