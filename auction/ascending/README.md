## Decentralized ascending auction

The ascending auction sample uses Hyperledger Fabric to run a decentralized auction. Two organizations have members that are bidders and sellers for a homogeneous good. Members submit their bids and asks to each organizations, who store them privately on their peers. The two organizations run an ascending price auction to allocate the goods between the bidders and sellers who have submitted bids.

Instead of operating under the guidance of a central auctioneer, each auction participant acts on behalf of the interest of the sellers who are members of their organization. Each organization tries to sell as many goods as possible for the highest possible price. The smart contract and the design of the auction allow each organization to run and close the auction at the equilibrium price while following their best interest.

## Auction design

Users who are interested in selling or buying a homogeneous good submit asks or bids to an organization that is a members of a Fabric network. Each bid or ask consists of a price and a quantity that the user is willing purchase or sell the good. Bids can be created or deleted only by their owner. However, each bid can be read by an auction administrator from the organization, who can then submit the bid or ask to an active action on the users behalf. Auctions can be started by user, and then can be join by any user as well. This scenario allows users to submit  preferences that are executed quickly when goods are brought to sale, such as goods moving through a supply chain or electricity entering the distribution grid.

Each auction is run as a series of rounds with a given price. Bid and asks are added to each round at a given price, announcing that the user is willing by buy or sell a given quantity at the round price. If the supply is greater than demand, a new round is created with the price rising by a set increment. Users or the auction administrator can then submit their bids and asks to the new round. When the demand is greater than supply and a new round does not need to be created, the auction is closed and the goods can be allocated. While the user or auction administrator can read the original bid or ask price, the original price is not revealed in the public auction. This allows users to keep some of their information private, and only reveal some of their preferences.

|  **Round** | **Price** | **Bids** |**Bids** |
| -----------|-----------|---------|---------|
| 2 | 40 | bidder1 - 20, bidder2 - 20, bidder3 - 20  | seller1 - 20, seller2 - 20, seller3 - 20|
| 1 | 30 | bidder1 - 20, bidder2 - 20, bidder3 - 20 | seller1 - 20, seller2 - 20|
| 0 | 20 | bidder1 - 20, bidder2 - 20, bidder3 - 20, bidder, bidder4 - 20 | seller1 - 20|
*Example 1: Each auction consists of multiple rounds with the price raised by a set increment.*

Auction can be started by any user who has access to the Fabric network. In addition to adding their bid or ask, any user can try to create a new round and raise the auction price. This allows sellers participating in the auction to try to raise the auction price. Any user can also close a round of the auction to stop new rounds from being created and set the final price of the auction.

While any user can try to update the auction, auction smart contract provides for techinical and economic mechanisms that prevent users from manipulating the auction. These mechanisms allow organizations that are competing to sell the most goods at the highest price to reach the an equilibrium where goods are allocated at a market clearing price.


## Technical mechanisms to prevent manipulation

The ascending auction smart contract is meant to be deployed with an endorsement policy of all the members of the channel that are participating in the auction. As a result, both organizations need to endorse any updates to the auction. As a result, both organizations need to agree to a series of checks before adding bids, creating new rounds, or closing the auction.

### Creating a new auction

Before a new auction is created, both organizations query their private data collection to check the lowest ask from their organization. The open price of the auction (the reserve price) needs to be lower than asks from any participating organization.

### Creating new round

Both organizations need to agree to a new round to raise the auction price. Before endorsing the creation of a new round, each organization checks:
- There is a previous round, and than a new round does not already exist
- The Demand of bids of the previous round is greater than supply of asks.

While all sellers have the incentive to try to create new rounds to increase supply, new rounds can only be created when there are more bids than asks for a given price.

### Closing a round

Both organizations need to agree that the auction has reached a market clearing price, and that the final round can be closed. Each auction runs a series of checks before they close an auction round:

- The supply of asks is greater than the demand of bids/
- All asks that are below the round price have been added to the round
- All bids that are above the round price have been added to the round
- Run the same set of checks on all any higher rounds.

These series of checks ensure that the auction cannot be closed while users are still adding bids and asks to the auction. This ensures that either organization cannot use the dynamic nature of the auction to end the auction at an artificially low or high price.

### Adding bids and asks

Both organizations need to approve the addition of any bids or asks to an auction round. This allows all participants from preventing uses from manipulating their bids. To see how bids and asks can be manipulated during the auction, see the scenario below.

|  **Round** | **Price** | **Bids** |**Bids** |
| -----------|-----------|---------|---------|
| 2 | 40 | bidder1 - 20, bidder2 - 20, bidder3 - 20  | seller1 - 20, seller2 - 20, seller3 - 20|
| 1 | 30 | bidder1 - 20, bidder2 - 20, bidder3 - 20 | seller1 - 20, seller2 - 20|
| 0 | 20 | bidder1 - 20, bidder2 - 20, bidder3 - 20, bidder, bidder4 - 20 | seller1 - 20|
*Example 2: Each auction consists of multiple rounds with the price raised by a set increment.*

In the example auction above, bidder 3 values the good at 40 dollars, meaning that he will earn exactly zero consumer surplus with the current auction. If the bidder lower their quantity to 10, the auction would not progress to round 2, and the price would remain at 30. By lowering the his bid, the bidder is able to change the auction equilibrium and ends up better off. Each organization enforces a set of rules to prevent users from bidding strategically:

- All bidders need to join the first round in order to join later rounds.
- All asks are automatically added to future rounds.
- Each organization checks the bid that is added to the auction against the hash on the public orderer book. This prevents users from altering their bids or asks while the auction is running.


## Economic incentive to prevent manipulation

The ascending auction also contains rules that are meant to align the incentives of the two organizations. In the example below,


|  **Round** | **Price** | **Bids** |**Asks** |
| -----------|-----------|---------|---------|
| 2 | 40 | bidder1 - 20, bidder2 - 20, bidder3 - 20  | seller1 - 20, seller2 - 20, seller3 - 20|
| 1 | 30 | bidder1 - 20, bidder2 - 20, bidder3 - 20 | seller1 - 20, seller2 - 20|
| 0 | 20 | bidder1 - 20, bidder2 - 20, bidder3 - 20, bidder, bidder4 - 20 | seller1 - 20|

To prevent different organizations from preferring different rounds, or from preferring a lower price, each ask os assigned the quantity that they bid for when the demand is in excess of supply, and keeps that quantity in the next round. In the example above, seller1 sells 20 goods, and is assigned that quantity for the rest of the action. The sum of all quantities assigned to each seller is the total quantity sold. The price received by each seller is set when the total quantity is greater than demand. By unlinking the quantity bid and the price awarded to their bid, each bidder maximizes their expected return by bidding truthfully.

|  **Round** | **Price** | **Bids** | **Asks** | **Sold** |
| -----------|-----------|---------|---------|--------|
| 2 | 40 | bidder1 - 20, bidder2 - 20, bidder3 - 20  | seller1 - 20, seller2 - 20, seller3 - 20| 60 |
| 1 | 30 | bidder1 - 20, bidder2 - 20, bidder3 - 20 | seller1 - 20, seller2 - 20| 40 |
| 0 | 20 | bidder1 - 20, bidder2 - 20, bidder3 - 20, bidder, bidder4 - 20 | seller1 - 20|20 |



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
./network.sh deployCC -ccn auction -ccp ../auction/ascending/chaincode-go/ -ccl go -ccep "AND('Org1MSP.peer','Org2MSP.peer')"
```

## Run the auction as Org1 and Org2

After you deployed the auction smart contract,

### Install the application dependencies

We provide as set of applications that you can use to run an example auction scenario. Change into the `application-javascript` directory:
```
cd fabric-samples/auction/ascending/application-javascript
```

From this directory, run the following command to download the application dependencies if you have not done so already:
```
npm install
```

### Register users and submit bids and asks

Before we can run the auction, we need to create the users who will participate in the auction, as well as the auction administrators that belong to Org1 and Org2. The users then need to submit bids and asks for the good before the auction can be run.

You can run the `auctionSetup.js` application to register the users that will participate in the auction. The bidder and sellers identities that are registered will add bids and asks for generic tickets to the respective private data collections.
```
node auctionSetup.js
```

The `auctionSetup.js` program provides an example set of bids and asks that we will use to run the auction. However, you can run your own scenario by editing the program and updating the bids and asks submitted to the auction,


### Start the auction

To run the auction, we will need to start two applications that are run by auction administrators from Org1 and Org2. The two administrators will act on behalf of the bidders and sellers who are members of their organization to submit their bids to any active auction.

Open two new terminals, one for the Org1 application, and another for Org2. In the Org1 terminal, make sure that you are in the `ascending/application-javascript` directory:
```
cd fabric-samples/auction/ascending/application-javascript
```

Run the following command to start the Org1 application:
```
node appOrg1.js
```

The application starts a listener that waits for an auction to be started. You can see the application notify you when the listener has started:
```
<-- added contract listener
```

In your terminal for Org2, make sure that you are the same directory:

```
cd fabric-samples/auction/ascending/application-javascript
```

You can then start the Org2 application:
```
node appOrg2.js
```

The Org2 application adds the same listener, but will then start an auction for tickets to sell. Once the auction starts, each organization runs a program to run the auction in a way that sells the most tickers for the highest possible price:

- Each application queries all the bids from their members to check if if their price is above the auction round. All bids above the price are added to the round.
- Each application also queries the asks from members that are sellers. Asks that are below the round price are added to the auction.
- Each organization will attempt to raise the price by creating a new round if the quantity demanded by the bidders is greater than the quantity sold.
- Each organization will close the auction only when the quantity sold is the same as demand.

Each organization starts running this logic when they learn that an auction has been started for a good that their members are interested in buying or selling. The mechanisms used by both organization ensure that both organizations can agree on a final price and close the auction.

After the auction is created, you can see each application generating logs in their respective terminal. You will see organization adding bids, asks, and trying to create new rounds. When both organizations agree to close the auction, you can see each application print the final auction round:
```
JSON Here
```

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
