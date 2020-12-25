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
*** Result ***SAVE THIS VALUE*** BidID: 4aacf2c25c09abfece161a593f0db9760ddee89539fc46a64f6829a07df514da
```

The BidID acts as the unique identifier for the bid. This ID allows you to query the bid using the `queryBid.js` program and add the bid to the auction. Save the bidID returned by the application as an environment variable in your terminal:
```
export BIDDER1_BID_ID=deca8ada788320d0b8a6f1ad41da148c8358b9236e5d6e6aa62d41a2f4c7b9b6
```

### Bid as bidder2

Let's submit another bid. Bidder2 would like to purchase 40 tickets for 50 dollars.
```
node bid.js org1 bidder2 tickets 20 40
```

Save the Bid ID returned by the application:
```
export BIDDER2_BID_ID=ad0c35265bd825ead7fdc8f76493a497aff2fb6d4dfc4646efe4281edd33a9fe
```


### Bid as bidder3 from Org2

Bidder3 will bid for 30 tickets at 70 dollars:
```
node bid.js org2 bidder3 tickets 20 60
```

Save the Bid ID returned by the application:
```
export BIDDER3_BID_ID=b7f76d2c601f684e489322ba1bd84ac2ca21e23ecb7db2e0e0b1f9560f0de87e
```


### Bid as bidder4

Bidder4 from Org2 would like to purchase 15 tickets for 60 dollars:
```
node bid.js org2 bidder4 tickets 20 80
```

Save the Bid ID returned by the application:
```
export BIDDER4_BID_ID=064fe070ed48d9f31418598ce5a17ecea6e5c41b6f639251552c9bfcfa33c4c1
```


### Bid as bidder5

Bidder5 from Org2 will bid for 20 tickets at 60 dollars:
```
node bid.js org2 bidder5 tickets 20 100
```

Save the Bid ID returned by the application:
```
export BIDDER5_BID_ID=ed51787bb9f57152ac745c0b81ce7354390c4df7de62a0292de0b7e299f61b7d
```

### Ask as seller1

Bidder1 will create a bid to purchase 50 tickets for 80 dollars.
```
node ask.js org1 seller1 tickets 20 30
```

The `ask.js` application also prints the bidID:
```
*** Result ***SAVE THIS VALUE*** AskID: 7bcecbeb3063214462e2a59a442d272eb97ebf03b51d5dfe06f2fd67db110eae
```

The BidID acts as the unique identifier for the bid. This ID allows you to query the bid using the `queryBid.js` program and add the bid to the auction. Save the bidID returned by the application as an environment variable in your terminal:
```
export SELLER1_BID_ID=7bcecbeb3063214462e2a59a442d272eb97ebf03b51d5dfe06f2fd67db110eae
```

### Ask as seller2

Let's submit another bid. Bidder2 would like to purchase 40 tickets for 50 dollars.
```
node ask.js org1 seller2 tickets 20 50
```

Save the Bid ID returned by the application:
```
export SELLER2_BID_ID=de32f14f01ed9795f2897fd8033ef76518318e709576bcc0c913e13c6eff940d
```


### Ask as seller3 from Org2

Bidder3 will bid for 30 tickets at 70 dollars:
```
node ask.js org2 bidder3 tickets 20 70
```

Save the Bid ID returned by the application:
```
export SELLER3_BID_ID=abe5dbf66b76ef40d0556830382120f4e9a55a98cb7530e051a2bc2a8a1816fa
```


### Ask as seller 4

Bidder4 from Org2 would like to purchase 15 tickets for 60 dollars:
```
node ask.js org2 bidder4 tickets 20 90
```

Save the Bid ID returned by the application:
```
export SELLER4_BID_ID=8fd985f6c4ef8f65546747e1eb16fac754ad3403d4ffca5643508d0d1e08fdcd
```

## Just the bids

```
export BIDDER1_BID_ID=5ed93b135bc45b598ecc63454a07efc419fc5249dab2c1bb86dc12726dbdb39b
export BIDDER2_BID_ID=1a856cdaba5d3ddcbaacd8d2eb5da9821f258e9112701f4a46c250371d34b3be
export BIDDER3_BID_ID=c9a72019cd0f047aa3db2a5f92f4d47cf2664bc50acbbda5b816022d559dd899
export BIDDER4_BID_ID=9a1b7f97441fbeebb9ac52ba31ffda3d114f03dc8493db187dfdda275d8944e8
export BIDDER5_BID_ID=49b8b5d9a55f22fc9c761bc2d86500a1cf734e46456528f8a2d998e0c1c9d767
export SELLER1_BID_ID=fd5d12bfb60f27ad5615a05ed84d7b1bd204043b1bbce49361608e67cdf4f195
export SELLER2_BID_ID=335ff48009a13ec2fd7d28955b587c40f01ad31042f4026ee55b80971f3ccd96
export SELLER3_BID_ID=957cd913b9ecb9f5c54ac33391d60db80a632376dcdf13ba509a532f65e79150
export SELLER4_BID_ID=86cbe68a1ce315ce15fa4cda41b3bf02d4681aeefb9f0e2b2bc6a278a5fe1688
```

## Create the auction


```
node createAuction.js org1 seller1 auction1 tickets 30
```

## Create auction round

```
node createNewRound.js org1 seller1 auction1 1 30
```

## add bids and asks


```
node submitBid.js org1 bidder1 auction1 0 20 $BIDDER1_BID_ID
```



```
node submitAsk.js org1 seller1 auction1 0 20 $SELLER1_BID_ID
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
```

The auction allocates tickets to the highest bids first. Because all 100 tickets are sold after allocating tickets to the bids that were submitted at 60, 60 is the `"price"` that clears the auction. The first 80 tickets are allocated to Bidder1 and Bidder3. The remaining 20 tickers are allocated to Bidder4 and Bidder5. When bids are tied, the auction smart contract fills the smaller bids first. As a result, Bidder4 is awarded their full bid of 15 tickets, while Bidder5 is allocated the remaining 5 tickets.
