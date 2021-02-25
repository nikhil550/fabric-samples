## Ascending auction

This example auction uses Hyperledger Fabric to run a decentralized auction. Instead of being run by a central auctioneer, the auction is run by organizations that represent bidders and sellers for the good being sold. Each organization acts in the interest of the sellers it represents, which fill as many asks as possible for the highest possible price. Even though the auction is not being run by an intermediary between competing organizations, the auction smart contract contains technical and economic mechanisms that prevent participants from manipulating the auction. The result of the competitive auction is the items are sold at an efficient market clearing price.

The auction is implemented as an ascending price auction that is run over multiple rounds. Buyers and sellers create bids and asks for a homogeneous good. After the auction is created, users can submit a bid or ask. If demand is greater than the supply being sold, a new round of the auction is created with a higher price. The auction stops when there is sufficient quantity to meet demand. The auction is designed to sell goods quickly against pre-existing bids and asks, such as when energy is supplied to the grid or goods moving through a supply chain.

This tutorial discusses how the auction is designed, as well as the technical and economic mechanisms the prevent users from manipulating the auction while it is running. You can then deploy the ascending auction smart contract to a running Fabric network to run an example auction. The auction is run using two applications that belong to separate organizations. The two applications interact with the auction in parallel and work to close the auction at an efficient price without cooperating.

- [Auction design](#auction-design)
- [Technical mechanisms to prevent manipulation](#technical-mechanisms-to-prevent-manipulation)
- [Economic mechanisms to prevent manipulation](#economic-mechanisms-to-prevent-manipulation)
- [Get started: Deploy the ascending auction smart contract](#get-started-deploy-the-ascending-auction-smart-contract)
- [Run the auction as Org1 and Org2](#get-started-deploy-the-ascending-auction-smart-contract)

## Auction design

Members of the Fabric network submit bids or asks to buy or sell a homogenous good. Each bid (or ask) specifies the quantity of the good that the user is willing to buy (or sell) at a given price. Bids can only be created or removed by user that will buy the good. Bids are not unique to each auction. Bids are created before an auction has started, and can be added to any auction that sells the bid is created for.

Each auction consists of a series of rounds. Each round has a set price. Bids and asks are added to each round, announcing that the user is willing to buy or sell the good at the bid or ask quantity at the round price. When quantity demanded by the bids is greater than the supply provided by the asks, a new round is created with the price raised by a set increment. When quantity is sufficient to meet supply, the auction is closed and the goods are allocated from the sellers to the buyers.

|  Round | Price | Bids:<br />Bidder - Quantity |Asks:<br />Seller - Quantity |
| -----------|-----------|---------|---------|
| 2 | 20 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20  | seller1 - 20<br />seller2 - 20<br />seller3 - 20|
| 1 | 15 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20 | seller1 - 20<br />seller2 - 20|
| 0 | 10 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20<br />bidder4 - 20 | seller1 - 20|

*Example 1: Each auction is composed of multiple rounds. Each round has a price that is raised by a set increment when a new round is created. Bidders and Sellers submit their bids or asks to each round.*

Bids and asks are stored in the private data collections of the organizations participating in the auction. The price at which users are willing to buy or sell is not shared with other organizations. Bids can be created or removed by the user that is willing to buy or sell. However, bids or asks can be added to an auction by an auction administrator from the users organization.
Each administrator is identified by an attribute added to a users certificate, whose access to users bids is governed by attribute based access control. Users only reveal the quantity that they are willing to buy or sell at a given price, without revealing their full bids.

Auctions can be started by any members of the organizations that are participants in the network. In addition to adding their bid or ask, any user can try to create a new round and raise the auction price. Users can also try to close the auction to try to set the final price of the auction. Any time an auction is started, the network becomes a non-cooperative game, with sellers try to create new rounds to increase the price and buyers trying to close the round to buy goods at the lowest possible price.

## Technical mechanisms to prevent manipulation

Because the ascending auction is meant to be run without a central auctioneer, the auction needs to prevent participants from altering the auction

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

|  Round | Price | Bids:<br />Bidder - Quantity |Asks:<br />Seller - Quantity |
| -----------|-----------|---------|---------|
| 2 | 20 | bidder1 - 20<br />bidder2 - 10<br />bidder3 - 10  | seller1 - 20<br />seller2 - 20<br />seller3 - 20|
| 1 | 15 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 10 | seller1 - 20<br />seller2 - 20|
| 0 | 10 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20<br />bidder4 - 20 | seller1 - 20|

*Example 2: Each auction consists of multiple rounds with the price raised by a set increment.*

In the example auction above, bidder 3 values the good at 40 dollars, meaning that he will earn exactly zero consumer surplus with the current auction. If the bidder lower their quantity to 10, the auction would not progress to round 2, and the price would remain at 30. By lowering the his bid, the bidder is able to change the auction equilibrium and ends up better off. Each organization enforces a set of rules to prevent users from bidding strategically:

- All bidders need to join the first round in order to join later rounds.
- All asks are automatically added to future rounds.
- Each organization checks the bid that is added to the auction against the hash on the public orderer book. This prevents users from altering their bids or asks while the auction is running.


## Economic mechanisms to prevent manipulation

The ascending auction also contains rules that are meant to align the incentives of the two organizations. In the example below,

|  Round | Price | Bids:<br />Bidder - Quantity |Asks:<br />Seller - Org - Quantity |
| -----------|-----------|---------|---------|
| 2 | 20 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 10  | seller1 - Org1 - 20<br />seller2 - Org1 - 20<br />seller3 - Org2 - 20|
| 1 | 15 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20 | seller1 - Org1 - 20<br />seller2 - Org1 - 20|
| 0 | 10 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20<br />bidder4 - 20 | seller1 - Org1 - 20|

*Example 3: Each auction consists of multiple rounds with the price raised by a set increment.*

To prevent different organizations from preferring different rounds, or from preferring a lower price, each ask os assigned the quantity that they bid for when the demand is in excess of supply, and keeps that quantity in the next round. In the example above, seller1 sells 20 goods, and is assigned that quantity for the rest of the action. The sum of all quantities assigned to each seller is the total quantity sold. The price received by each seller is set when the total quantity is greater than demand. By unlinking the quantity bid and the price awarded to their bid, each bidder maximizes their expected return by bidding truthfully.

|  Round | Price | Bids:<br />Bidder - Quantity |Asks:<br />Seller - Org - Quantity | Quantity sold |
| -----------|-----------|---------|---------|--------|
| 2 | 40 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 10  | seller1 - Org1 - 20<br />seller2 - Org1 - 20<br />seller3 - Org2 - 20| 50 |
| 1 | 30 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20 | seller1 - Org1 - 20<br />seller2 - Org1 - 20| 40 |
| 0 | 20 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20<br />bidder4 - 20 | seller1 - Org1 - 20| 20 |

*Example 4: Each auction consists of multiple rounds with the price raised by a set increment.*

## Get started: Deploy the ascending auction smart contract

Now that we understand how the auction works, we can deploy the ascending auction smart contract to the Fabric test network to run an example auction. The Fabric test network contains two organizations, Org1 and Org1, that will act as the trust anchors for the auction. Each organization runs one peer that will store the bids and asks from their members.

If you have not already, follow the instructions to blah blah blah.

Clone this repo.

Once you have cloned the repo, change into the test network directory.
```
cd fabric-samples/test-network
```

If the test network is already running, run the following command to bring the network down and start from a clean initial state.
```
./network.sh down
```

You can then run the following command to deploy a new network. We will also deploy a Certificate Authority for each organization that we will use to create the auction administrator, bidders, and sellers from each organization.
```
./network.sh up createChannel -ca
```

Run the following command to deploy the auction smart contract.
```
./network.sh deployCC -ccn auction -ccp ../auction/ascending/chaincode-go/ -ccl go -ccep "AND('Org1MSP.peer','Org2MSP.peer')"
```

## Run the auction as Org1 and Org2

After you deployed the auction smart contract, we can run the auction using a set of Node.js applications. Change into the `application-javascript` directory:
```
cd fabric-samples/auction/ascending/application-javascript
```

From this directory, run the following command to download the application dependencies:
```
npm install
```

### Register users and submit bids and asks

Before we can run the auction, we need to create the seller and bidder identities from Org1 and Org1, as well as the auction administrators that we will use to run the auction on those users behalf. The users then need to submit bids and asks for the good before the auction can be run.

You can run the `auctionSetup.js` application to register the users that will participate in the auction. The bidder and seller identities that are registered will add bids and asks for generic tickets to the respective private data collections.
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
