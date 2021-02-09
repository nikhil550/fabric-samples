## Decentralized ascending auction

The ascending auction sample uses a Hyperledger Fabric smart contract to run a decentralized auction. Two organizations, acting on behalf of sellers and bidders submitting bids and asks to their organization, participate in an ascending double auction. The price of the auction will rise until quantity sold is equal to demand. Instead of operating under the guidance of a central auctioneer, each organization acts in the self of interest of the sellers who submitted, and try to drive the price of the auction as high as possible. Bidders can participate in the auction can try to close the auction at the lowest possible price. However, the Fabric smart contract model will only allow the auction to be closed if both organizations agree that the auction has reached the equilibrium price.

## Auction scenario

The auction is designed to facilitate the quick sale of goods between a set of buyers and sellers, such as energy entering the distribution grid or goods moving through a supply chain.


## Technical incentive compatability


## Economic incentive compatability




## Deploy the ascending auction smart contract

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
./network.sh deployCC -ccn auction -ccp ../auction/ascending/chaincode-go/ -ccl go
```

## Install the application dependencies

We provide as set of applications that you can use to run an example auction scenario. Change into the `application-javascript` directory:
```
cd fabric-samples/auction/ascending/application-javascript
```

From this directory, run the following command to download the application dependencies if you have not done so already:
```
npm install
```

## Setup the users, bids and asks

Before we can run the auction, we need to create the users who will participate in the auction, as well as the auction administrators that belong to Org1 and Org2. The users then need to submit bids and asks for the good before the auction can be run.

You can run the `auctionSetup.js` application to register the users that will participate in the auction. The bidder and sellers identities that are registered will add bids and asks for generic tickets to the respective private data collections.
```
node auctionSetup.js
```

The `auctionSetup.js` program provides an example set of bids and asks that we will use to run the auction. However, you can run your own scenario by editing the program and updating the bids and asks submitted to the auction,


## Run the auction

To run the auction, we will need to start two applications that are run by auction administrators from Org1 and Org2. The two administrators will act on behalf of the bidders and sellers who are members of their organization to submit their bids to any active auction.

Open two new terminals, one for the Org1 application, and another for Org2. In the Org1 terminal, make sure that you are in the `ascending/application-javascript` directory:
```
cd fabric-samples/auction/ascending/application-javascript
```

Run the following command to start the Org1 application:
```
node appOrg1.js
```

The main component of the application is a listener that waits for an auction to be started. You can see the application notify you when the auction has started:
```
<-- added contract listener
```

When an auction starts, the action searches his peer for bids and asks that have been issued for the same item, and starts the logic for adding bids and asks to the auction and starting new rounds.

In your terminal for Org2, make sure that you are the same directory:

```
cd fabric-samples/auction/ascending/application-javascript
```

You can then start the Org2 application:
```
node appOrg2.js
```

The Org2 application adds the same listener, but will then start an auction for tickets. After the auction is created, you can see each application generating logs in their respective terminal.

## Clean up

When your are done using the auction smart contract, you can bring down the network and clean up the environment. In the `application-javascript` directory, run the following command to remove the wallets used to run the applications:
```
rm -rf wallet
```

You can then navigate to the test network directory and bring down the network:
````
cd ../../test-network/
./network.sh down
````
