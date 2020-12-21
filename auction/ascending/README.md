## Ascending auction

Test readme for an ascending auction.

## Deploy the chaincode

Change into the test network directory.
```
cd fabric-samples/test-network
```

If the test network is already running, run the following command to bring the network down and start from a clean initial state.
```
./network.sh down
```

You can then run the following command to deploy a new network.
```
./network.sh up createChannel -ca
```

Run the following command to deploy the auction smart contract.
```
./network.sh deployCC -ccn auction -ccp ../auction/ascending/chaincode-go/ -ccep "OR('Org1MSP.peer','Org2MSP.peer')" -ccl go
```

## Install the application dependencies

Change into the `application-javascript` directory:
```
cd fabric-samples/auction/ascending/application-javascript
```

From this directory, run the following command to download the application dependencies if you have not done so already:
```
npm install
```

## Register and enroll the application identities

To interact with the network, you will need to enroll the Certificate Authority administrators of Org1 and Org2. You can use the `enrollAdmin.js` program for this task. Run the following command to enroll the Org1 admin:
```
node enrollAdmin.js org1
```
You should see the logs of the admin wallet being created on your local file system. Now run the command to enroll the CA admin of Org2:
```
node enrollAdmin.js org2
```

We can use the CA admins of both organizations to register and enroll the identities of the seller that will create the auction and the bidders who will try to purchase the tickets.


You should see the logs of the seller wallet being created as well. Run the following commands to register and enroll 2 bidders from Org1 and another 3 bidders from Org2:
```
node registerEnrollUser.js org1 bidder1
node registerEnrollUser.js org1 bidder2
node registerEnrollUser.js org2 bidder3
node registerEnrollUser.js org2 bidder4
node registerEnrollUser.js org2 bidder5
node registerEnrollUser.js org1 seller1
node registerEnrollUser.js org1 seller2
node registerEnrollUser.js org2 seller3
node registerEnrollUser.js org2 seller4
```

## Bid on the items

We can now use the bidder wallets to submit bids to the auction:

### Bid as bidder1

Bidder1 will create a bid to purchase 50 tickets for 80 dollars.
```
node bid.js org1 bidder1 tickets 20 20
```

The `bid.js` application also prints the bidID:
```
*** Result ***SAVE THIS VALUE*** BidID: b38ef6a2a0ef0d15df5937331081e52d46f5d3aa75c652eec7e9185ff487eee0
```

The BidID acts as the unique identifier for the bid. This ID allows you to query the bid using the `queryBid.js` program and add the bid to the auction. Save the bidID returned by the application as an environment variable in your terminal:
```
export BIDDER1_BID_ID=b38ef6a2a0ef0d15df5937331081e52d46f5d3aa75c652eec7e9185ff487eee0
```

### Bid as bidder2

Let's submit another bid. Bidder2 would like to purchase 40 tickets for 50 dollars.
```
node bid.js org1 bidder2 tickets 20 40
```

Save the Bid ID returned by the application:
```
export BIDDER2_BID_ID=0bdf27e7107dab731dcffebf38163c3008ff61aad0ab864f36ae2f3cb7384a7d
```


### Bid as bidder3 from Org2

Bidder3 will bid for 30 tickets at 70 dollars:
```
node bid.js org2 bidder3 tickets 20 60
```

Save the Bid ID returned by the application:
```
export BIDDER3_BID_ID=c461e2df48e71d2f2376d6fe0fe512e207af8388280b9ad578f693425ec25b21
```


### Bid as bidder4

Bidder4 from Org2 would like to purchase 15 tickets for 60 dollars:
```
node bid.js org2 bidder4 tickets 20 80
```

Save the Bid ID returned by the application:
```
export BIDDER4_BID_ID=7dab4e516eda9898b7290a79e841dc4c12d946ae76261bba37241347c6ca2dc7
```


### Bid as bidder5

Bidder5 from Org2 will bid for 20 tickets at 60 dollars:
```
node bid.js org2 bidder5 tickets 20 100
```

Save the Bid ID returned by the application:
```
export BIDDER5_BID_ID=604e8ccbd764687143ccf3312dd135d95914bed3e59a3235b14b3b6574fc63e9
```

### Ask as seller1

Bidder1 will create a bid to purchase 50 tickets for 80 dollars.
```
node ask.js org1 seller1 tickets 20 30
```

The `ask.js` application also prints the bidID:
```
*** Result ***SAVE THIS VALUE*** BidID: 8158c009c109610ae457ebd082b512abcb002944ce7909896d10e2396a0a4903
```

The BidID acts as the unique identifier for the bid. This ID allows you to query the bid using the `queryBid.js` program and add the bid to the auction. Save the bidID returned by the application as an environment variable in your terminal:
```
export SELLER1_BID_ID=8158c009c109610ae457ebd082b512abcb002944ce7909896d10e2396a0a4903
```

### Ask as seller2

Let's submit another bid. Bidder2 would like to purchase 40 tickets for 50 dollars.
```
node ask.js org1 seller2 tickets 20 50
```

Save the Bid ID returned by the application:
```
export SELLER2_BID_ID=c5002120f2725d228840b733608597e90708ebeeeed53c0aa0a39d37032cd783
```


### Ask as seller3 from Org2

Bidder3 will bid for 30 tickets at 70 dollars:
```
node ask.js org2 bidder3 tickets 20 70
```

Save the Bid ID returned by the application:
```
export SELLER3_BID_ID=a00cdf3c450f0aa7b7535bc9d6c9ba3d043cad0435a4d46a3fe862875642621e
```


### Ask as seller 4

Bidder4 from Org2 would like to purchase 15 tickets for 60 dollars:
```
node ask.js org2 bidder4 tickets 20 90
```

Save the Bid ID returned by the application:
```
export SELLER4_BID_ID=5d2de7d18c39c01268d37d8e18b4f99d81f3b16b60234b1925fc792bf147a0b6
```

## Just the bids

```
export BIDDER1_BID_ID=b38ef6a2a0ef0d15df5937331081e52d46f5d3aa75c652eec7e9185ff487eee0
```
```
export BIDDER2_BID_ID=b38ef6a2a0ef0d15df5937331081e52d46f5d3aa75c652eec7e9185ff487eee0
```
```
export BIDDER3_BID_ID=b38ef6a2a0ef0d15df5937331081e52d46f5d3aa75c652eec7e9185ff487eee0
```
```
export BIDDER4_BID_ID=b38ef6a2a0ef0d15df5937331081e52d46f5d3aa75c652eec7e9185ff487eee0
```
```
export BIDDER5_BID_ID=b38ef6a2a0ef0d15df5937331081e52d46f5d3aa75c652eec7e9185ff487eee0
```

```
export SELLER1_BID_ID=5d2de7d18c39c01268d37d8e18b4f99d81f3b16b60234b1925fc792bf147a0b6
```
```
export SELLER2_BID_ID=5d2de7d18c39c01268d37d8e18b4f99d81f3b16b60234b1925fc792bf147a0b6
```
```
export SELLER3_BID_ID=5d2de7d18c39c01268d37d8e18b4f99d81f3b16b60234b1925fc792bf147a0b6
```
```
export SELLER4_BID_ID=5d2de7d18c39c01268d37d8e18b4f99d81f3b16b60234b1925fc792bf147a0b6
```

## Create the auction

The seller from Org1 would like to create an auction to sell 100 tickets. Run the following command to use the seller wallet to run the `createAuction.js` application. The seller needs to provide an ID for the auction, the item to be sold, and the quantity to be sold to create the auction:
```
node createAuction.js org1 seller1 auction1 tickets 40
```

You will see the application query the auction after it is created.

```
node createNewRound.js org1 seller1 auction1 1 30
```

submit bid
```
node submitBid.js org1 bidder1 auction1 $BIDDER1_BID_ID
```

The hash of bid is added to the list of private bids in that have been submitted to `auction1`. Storing the hash on the public auction ledger allows users to prove the accuracy of the bids they reveal once bidding is closed. The application queries the auction to verify that the bid was added:
```
*** Result: Auction: {
  "objectType": "auction",
  "item": "tickets",
  "seller": "eDUwOTo6Q049c2VsbGVyLE9VPWNsaWVudCtPVT1vcmcxK09VPWRlcGFydG1lbnQxOjpDTj1jYS5vcmcxLmV4YW1wbGUuY29tLE89b3JnMS5leGFtcGxlLmNvbSxMPUR1cmhhbSxTVD1Ob3J0aCBDYXJvbGluYSxDPVVT",
  "quantity": 100,
  "organizations": [
    "Org1MSP"
  ],
  "privateBids": {
    "\u0000bid\u0000auction1\u000061e9b0fc1913f10872625bea4a6555522c70070416209848cc1d8fb6101133ad\u0000": {
      "org": "Org1MSP",
      "hash": "584dbad2269a44afb42bdbc7a7a4c08a7cd50deece1eb3fc38d4e49b9342f270"
    }
  },
  "revealedBids": {},
  "winners": [],
  "price": 0,
  "status": "open"
}
```



Submit bidder2's bid to the auction:
```
node submitBid.js org1 bidder2 auction1 $BIDDER2_BID_ID
```


Add bidder3's bid to the auction:
```
node submitBid.js org2 bidder3 auction1 $BIDDER3_BID_ID
```

Because bidder3 belongs to Org2, submitting the bid will add Org2 to the list of participating organizations. You can see the Org2 MSP ID has been added to the list of `"organizations"` in the updated auction returned by the application:
```
*** Result: Auction: {
  "objectType": "auction",
  "item": "tickets",
  "seller": "eDUwOTo6Q049c2VsbGVyLE9VPWNsaWVudCtPVT1vcmcxK09VPWRlcGFydG1lbnQxOjpDTj1jYS5vcmcxLmV4YW1wbGUuY29tLE89b3JnMS5leGFtcGxlLmNvbSxMPUR1cmhhbSxTVD1Ob3J0aCBDYXJvbGluYSxDPVVT",
  "quantity": 100,
  "organizations": [
    "Org1MSP",
    "Org2MSP"
  ],
  "privateBids": {
    "\u0000bid\u0000auction1\u000061e9b0fc1913f10872625bea4a6555522c70070416209848cc1d8fb6101133ad\u0000": {
      "org": "Org1MSP",
      "hash": "584dbad2269a44afb42bdbc7a7a4c08a7cd50deece1eb3fc38d4e49b9342f270"
    },
    "\u0000bid\u0000auction1\u0000911c7920a7ba4643a531cb2d5d274d303fdb2e6800f50aeb6b725af0b7162ea2\u0000": {
      "org": "Org1MSP",
      "hash": "bbcd0c7c376e6681a76d8c5482c97f8bdcda55c90c5478100c3aef17815c4fd3"
    },
    "\u0000bid\u0000auction1\u000093a8164628fa28290554b5dc6f505cbb8c7498d8f7c60f7df33d4a1cffb8fa47\u0000": {
      "org": "Org2MSP",
      "hash": "4446c7eb0e2d64165a916ee996348a18716f4c97e632d58d5a8c20eeec5a9238"
    }
  },
  "revealedBids": {},
  "winners": [],
  "price": 0,
  "status": "open"
}
```

Now that a bid from Org2 has been added to the auction, any updates to the auction need to be endorsed by the Org2 peer. The applications will use the `"organizations"` field to specify which organizations need to endorse submitting a new bid, revealing a bid, or updating the auction status.


Add bidder4's bid to the auction:
```
node submitBid.js org2 bidder4 auction1 $BIDDER4_BID_ID
```

Add bidder4's bid to the auction:
```
node submitBid.js org2 bidder5 auction1 $BIDDER5_BID_ID
```


## Close the auction

Now that all five bidders have joined the auction, the seller would like to close the auction and allow buyers to reveal their bids. The seller identity that created the auction needs to submit the transaction:
```
node closeAuction.js org1 seller auction1
```

The application will query the auction to allow you to verify that the auction status has changed to closed.

## Reveal bids

After the auction is closed, bidders can try to win the auction by revealing their bids. The transaction to reveal a bid needs to pass four checks:
1. The auction is closed.
2. The transaction was submitted by the identity that created the bid.
3. The hash of the revealed bid matches the hash of the bid on the channel ledger. This confirms that the bid is the same as the bid that is stored in the private data collection.
4. The hash of the revealed bid matches the hash that was submitted to the auction. This confirms that the bid was not altered after the auction was closed.

Use the `revealBid.js` application to reveal the bid of Bidder1:
```
node revealBid.js org1 bidder1 auction1 $BIDDER1_BID_ID
```

The full bid details, including the quantity and price, are now visible:
```
*** Result: Auction: {
  "objectType": "auction",
  "item": "tickets",
  "seller": "eDUwOTo6Q049c2VsbGVyLE9VPWNsaWVudCtPVT1vcmcxK09VPWRlcGFydG1lbnQxOjpDTj1jYS5vcmcxLmV4YW1wbGUuY29tLE89b3JnMS5leGFtcGxlLmNvbSxMPUR1cmhhbSxTVD1Ob3J0aCBDYXJvbGluYSxDPVVT",
  "quantity": 100,
  "organizations": [
    "Org1MSP",
    "Org2MSP"
  ],
  "privateBids": {
    "\u0000bid\u0000auction1\u00003245049bd81a8cecfbc006f29eed1df0570e385e63daedd19c06b2c92d2067ae\u0000": {
      "org": "Org2MSP",
      "hash": "5a51070f188e6480fe606acdc5ad1a1c36330adc3547eb021aaf87aa00ec79f7"
    },
    "\u0000bid\u0000auction1\u0000324de04c459c5a38f103e9096dea06e19faca88f84dd5175ba0e6fd6a9d7d140\u0000": {
      "org": "Org2MSP",
      "hash": "516c775f7da2fd653dab71d20d60a6e8bea8ff856803af4b59f943ca2ba40699"
    },
    "\u0000bid\u0000auction1\u000061e9b0fc1913f10872625bea4a6555522c70070416209848cc1d8fb6101133ad\u0000": {
      "org": "Org1MSP",
      "hash": "584dbad2269a44afb42bdbc7a7a4c08a7cd50deece1eb3fc38d4e49b9342f270"
    },
    "\u0000bid\u0000auction1\u0000911c7920a7ba4643a531cb2d5d274d303fdb2e6800f50aeb6b725af0b7162ea2\u0000": {
      "org": "Org1MSP",
      "hash": "bbcd0c7c376e6681a76d8c5482c97f8bdcda55c90c5478100c3aef17815c4fd3"
    },
    "\u0000bid\u0000auction1\u000093a8164628fa28290554b5dc6f505cbb8c7498d8f7c60f7df33d4a1cffb8fa47\u0000": {
      "org": "Org2MSP",
      "hash": "4446c7eb0e2d64165a916ee996348a18716f4c97e632d58d5a8c20eeec5a9238"
    }
  },
  "revealedBids": {
    "\u0000bid\u0000auction1\u000061e9b0fc1913f10872625bea4a6555522c70070416209848cc1d8fb6101133ad\u0000": {
      "objectType": "bid",
      "quantity": 50,
      "price": 80,
      "org": "Org1MSP",
      "buyer": "eDUwOTo6Q049YmlkZGVyMSxPVT1jbGllbnQrT1U9b3JnMStPVT1kZXBhcnRtZW50MTo6Q049Y2Eub3JnMS5leGFtcGxlLmNvbSxPPW9yZzEuZXhhbXBsZS5jb20sTD1EdXJoYW0sU1Q9Tm9ydGggQ2Fyb2xpbmEsQz1VUw=="
    }
  },
  "winners": [],
  "price": 0,
  "status": "closed"
}
```

All bidders will reveal their bid to participate in the auction. Run the following commands to reveal the bids of the remaining four bidders:
```
node revealBid.js org1 bidder2 auction1 $BIDDER2_BID_ID
node revealBid.js org2 bidder3 auction1 $BIDDER3_BID_ID
node revealBid.js org2 bidder4 auction1 $BIDDER4_BID_ID
node revealBid.js org2 bidder5 auction1 $BIDDER5_BID_ID
```

## End the auction

Now that the winning bids have been revealed, we can end the auction:
```
node endAuction org1 seller auction1
```

The transaction was successfully endorsed by both Org1 and Org2, who both calculated the same price and winners of the auction. Each winning bidder is listed next to the quantity that was allocated to them.

```
*** Result: Auction: {
  "objectType": "auction",
  "item": "tickets",
  "seller": "eDUwOTo6Q049c2VsbGVyLE9VPWNsaWVudCtPVT1vcmcxK09VPWRlcGFydG1lbnQxOjpDTj1jYS5vcmcxLmV4YW1wbGUuY29tLE89b3JnMS5leGFtcGxlLmNvbSxMPUR1cmhhbSxTVD1Ob3J0aCBDYXJvbGluYSxDPVVT",
  "quantity": 100,
  "organizations": [
    "Org1MSP",
    "Org2MSP"
  ],
  "privateBids": {
    "\u0000bid\u0000auction1\u00003245049bd81a8cecfbc006f29eed1df0570e385e63daedd19c06b2c92d2067ae\u0000": {
      "org": "Org2MSP",
      "hash": "5a51070f188e6480fe606acdc5ad1a1c36330adc3547eb021aaf87aa00ec79f7"
    },
    "\u0000bid\u0000auction1\u0000324de04c459c5a38f103e9096dea06e19faca88f84dd5175ba0e6fd6a9d7d140\u0000": {
      "org": "Org2MSP",
      "hash": "516c775f7da2fd653dab71d20d60a6e8bea8ff856803af4b59f943ca2ba40699"
    },
    "\u0000bid\u0000auction1\u000061e9b0fc1913f10872625bea4a6555522c70070416209848cc1d8fb6101133ad\u0000": {
      "org": "Org1MSP",
      "hash": "584dbad2269a44afb42bdbc7a7a4c08a7cd50deece1eb3fc38d4e49b9342f270"
    },
    "\u0000bid\u0000auction1\u0000911c7920a7ba4643a531cb2d5d274d303fdb2e6800f50aeb6b725af0b7162ea2\u0000": {
      "org": "Org1MSP",
      "hash": "bbcd0c7c376e6681a76d8c5482c97f8bdcda55c90c5478100c3aef17815c4fd3"
    },
    "\u0000bid\u0000auction1\u000093a8164628fa28290554b5dc6f505cbb8c7498d8f7c60f7df33d4a1cffb8fa47\u0000": {
      "org": "Org2MSP",
      "hash": "4446c7eb0e2d64165a916ee996348a18716f4c97e632d58d5a8c20eeec5a9238"
    }
  },
  "revealedBids": {
    "\u0000bid\u0000auction1\u00003245049bd81a8cecfbc006f29eed1df0570e385e63daedd19c06b2c92d2067ae\u0000": {
      "objectType": "bid",
      "quantity": 20,
      "price": 60,
      "org": "Org2MSP",
      "buyer": "eDUwOTo6Q049YmlkZGVyNCxPVT1jbGllbnQrT1U9b3JnMitPVT1kZXBhcnRtZW50MTo6Q049Y2Eub3JnMi5leGFtcGxlLmNvbSxPPW9yZzIuZXhhbXBsZS5jb20sTD1IdXJzbGV5LFNUPUhhbXBzaGlyZSxDPVVL"
    },
    "\u0000bid\u0000auction1\u0000324de04c459c5a38f103e9096dea06e19faca88f84dd5175ba0e6fd6a9d7d140\u0000": {
      "objectType": "bid",
      "quantity": 15,
      "price": 60,
      "org": "Org2MSP",
      "buyer": "eDUwOTo6Q049YmlkZGVyNCxPVT1jbGllbnQrT1U9b3JnMitPVT1kZXBhcnRtZW50MTo6Q049Y2Eub3JnMi5leGFtcGxlLmNvbSxPPW9yZzIuZXhhbXBsZS5jb20sTD1IdXJzbGV5LFNUPUhhbXBzaGlyZSxDPVVL"
    },
    "\u0000bid\u0000auction1\u000061e9b0fc1913f10872625bea4a6555522c70070416209848cc1d8fb6101133ad\u0000": {
      "objectType": "bid",
      "quantity": 50,
      "price": 80,
      "org": "Org1MSP",
      "buyer": "eDUwOTo6Q049YmlkZGVyMSxPVT1jbGllbnQrT1U9b3JnMStPVT1kZXBhcnRtZW50MTo6Q049Y2Eub3JnMS5leGFtcGxlLmNvbSxPPW9yZzEuZXhhbXBsZS5jb20sTD1EdXJoYW0sU1Q9Tm9ydGggQ2Fyb2xpbmEsQz1VUw=="
    },
    "\u0000bid\u0000auction1\u0000911c7920a7ba4643a531cb2d5d274d303fdb2e6800f50aeb6b725af0b7162ea2\u0000": {
      "objectType": "bid",
      "quantity": 40,
      "price": 50,
      "org": "Org1MSP",
      "buyer": "eDUwOTo6Q049YmlkZGVyMixPVT1jbGllbnQrT1U9b3JnMStPVT1kZXBhcnRtZW50MTo6Q049Y2Eub3JnMS5leGFtcGxlLmNvbSxPPW9yZzEuZXhhbXBsZS5jb20sTD1EdXJoYW0sU1Q9Tm9ydGggQ2Fyb2xpbmEsQz1VUw=="
    },
    "\u0000bid\u0000auction1\u000093a8164628fa28290554b5dc6f505cbb8c7498d8f7c60f7df33d4a1cffb8fa47\u0000": {
      "objectType": "bid",
      "quantity": 30,
      "price": 70,
      "org": "Org2MSP",
      "buyer": "eDUwOTo6Q049YmlkZGVyMyxPVT1jbGllbnQrT1U9b3JnMitPVT1kZXBhcnRtZW50MTo6Q049Y2Eub3JnMi5leGFtcGxlLmNvbSxPPW9yZzIuZXhhbXBsZS5jb20sTD1IdXJzbGV5LFNUPUhhbXBzaGlyZSxDPVVL"
    }
  },
  "winners": [
    {
      "buyer": "eDUwOTo6Q049YmlkZGVyMSxPVT1jbGllbnQrT1U9b3JnMStPVT1kZXBhcnRtZW50MTo6Q049Y2Eub3JnMS5leGFtcGxlLmNvbSxPPW9yZzEuZXhhbXBsZS5jb20sTD1EdXJoYW0sU1Q9Tm9ydGggQ2Fyb2xpbmEsQz1VUw==",
      "quantity": 50
    },
    {
      "buyer": "eDUwOTo6Q049YmlkZGVyMyxPVT1jbGllbnQrT1U9b3JnMitPVT1kZXBhcnRtZW50MTo6Q049Y2Eub3JnMi5leGFtcGxlLmNvbSxPPW9yZzIuZXhhbXBsZS5jb20sTD1IdXJzbGV5LFNUPUhhbXBzaGlyZSxDPVVL",
      "quantity": 30
    },
    {
      "buyer": "eDUwOTo6Q049YmlkZGVyNCxPVT1jbGllbnQrT1U9b3JnMitPVT1kZXBhcnRtZW50MTo6Q049Y2Eub3JnMi5leGFtcGxlLmNvbSxPPW9yZzIuZXhhbXBsZS5jb20sTD1IdXJzbGV5LFNUPUhhbXBzaGlyZSxDPVVL",
      "quantity": 15
    },
    {
      "buyer": "eDUwOTo6Q049YmlkZGVyNCxPVT1jbGllbnQrT1U9b3JnMitPVT1kZXBhcnRtZW50MTo6Q049Y2Eub3JnMi5leGFtcGxlLmNvbSxPPW9yZzIuZXhhbXBsZS5jb20sTD1IdXJzbGV5LFNUPUhhbXBzaGlyZSxDPVVL",
      "quantity": 5
    }
  ],
  "price": 60,
  "status": "ended"
}
```

The auction allocates tickets to the highest bids first. Because all 100 tickets are sold after allocating tickets to the bids that were submitted at 60, 60 is the `"price"` that clears the auction. The first 80 tickets are allocated to Bidder1 and Bidder3. The remaining 20 tickers are allocated to Bidder4 and Bidder5. When bids are tied, the auction smart contract fills the smaller bids first. As a result, Bidder4 is awarded their full bid of 15 tickets, while Bidder5 is allocated the remaining 5 tickets.

## Clean up

When your are done using the auction smart contract, you can bring down the network and clean up the environment. In the `auction/application-javascript` directory, run the following command to remove the wallets used to run the applications:
```
rm -rf wallet
```

You can then navigate to the test network directory and bring down the network:
````
cd ../../../test-network/
./network.sh down
````
