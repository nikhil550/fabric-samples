## Ascending auction

This example uses a Hyperledger Fabric smart contract to run a decentralized auction. Instead of being run by a central auctioneer, the auction is run by self interested organizations that are members of the Fabric network. Each organization has members that are buyers and sellers of a good that is being sold at auction. When running the auction, each organization attempts to fill as many asks from its sellers at the highest possible price. The smart contract contains technical and economic mechanisms that prevent organizations from manipulating the outcome to their benefit. As a result, the members can run a competitive auction that sells items at an efficient market clearing price without the presence of a central intermediary.

The smart contract implements an ascending price auction that can sell multiple items of the same type. The auction can be used by multiple sellers, as long as they are selling the same good. Buyers and sellers create bids for the item before the auction starts. Users can submit their bids and asks to the auction after it is created. The price of the auction rises until supply is sufficient to meet demand. Although the auction can be used for many situations, it is designed to allocate goods quickly against pre-existing bids and asks from users, such as energy entering the electricity grid or goods moving through a supply chain.

This tutorial discusses how the auction is implemented and the technical and economic mechanisms that protect the auction from manipulation by self interested users. You can run an example auction by deploying the ascending auction smart contract to a running Fabric network. To demonstrate the decentralized nature of the auction, you can run the auction by submitting asks and bids to the auction using two applications run by different organizations. The applications interact with the auction in parallel and are able to close the auction at an efficient price without coordination or cooperation.

- [Auction design](#auction-design)
- [Technical mechanisms to prevent manipulation](#technical-mechanisms-to-prevent-manipulation)
- [Economic mechanisms to prevent manipulation](#economic-mechanisms-to-prevent-manipulation)
- [Get started: deploy the ascending auction smart contract](#get-started-deploy-the-ascending-auction-smart-contract)
- [Run the auction as Org1 and Org2](#get-started-deploy-the-ascending-auction-smart-contract)

## Auction design

While the auction is run by organizations that are members of the Fabric network, the buyers and sellers who participate in the auction are individual users from those organizations. Users create a bid or ask to sell or buy and homogenous good. Each bid (or ask) specifies the quantity of the good that the user is willing to buy (or sell) at a given price. Only the buyer or seller of the good can create or remove the bid. Bids and asks are stored in the private data collections of the organizations participating in the auction. The price at which users are willing to buy or sell is not shared with other organizations. Bidders do not need to create a bid that is unique to each auction. Bids or asks can be added to any auction that sells the item.

Each auction consists of multiple rounds. Each round has a price that the good can be bought or sold for. users can add their bids or asks to each round, announcing the quantity of the good that they are willing to buy or sell at the round price. When quantity demanded by bids is greater than the supply provided by asks, a new round of the auction is created. The new round raises the price by a set increment, and allows buyers and sellers to submit new bids. Users only reveal the quantity that they are willing to buy or sell at a given price, without revealing their full bids. When the quantity for a round is sufficient to meet supply, the auction is closed and the goods are allocated from the sellers to the buyers at the final round price.

|  Round | Price | Bids:<br />Bidder - Quantity |Asks:<br />Seller - Quantity |
| -----------|-----------|---------|---------|
| 2 | 20 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20  | seller1 - 20<br />seller2 - 20<br />seller3 - 20|
| 1 | 15 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20 | seller1 - 20<br />seller2 - 20|
| 0 | 10 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20<br />bidder4 - 20 | seller1 - 20|

*Example 1: Each auction consists of multiple rounds. The price of each round is raised by a set increment when a new round is created. Bids and asks are added separately to each round.*

Auctions can be started by any users from the organizations that are network participants. Bids or asks can be added to the auction either by the creator or an auction administrator from the users organization. Auction administrators are differentiated from other participants using attribute based access control. In addition to adding their bid or ask, any user can try to create a new round and raise the auction price. Any uses can also try to close the auction to set the final price. Any time an auction is created, the auction becomes a non-cooperative game between buyers and sellers. Sellers can try to create new rounds to raise the auction price while buyers try to close the auction and purchase goods for the lowest possible price.


## Technical mechanisms to prevent manipulation

Because the ascending auction smart contract is meant to be run without a central auctioneer, the auction needs to be protected against from self interested organizations who have an interest in altering the auction outcome. The smart contract and the Fabric blockchain work together to prevent organizations from manipulating the auction.

### Smart contract endorsement policy

Assets that are stored on the Fabric blockchain ledger are protected by the smart contract endorsement policy. To update an entry on the ledger, A sufficient number of organizations need to agree to the update to meet the smart contract endorsement policy.  When the ascending auction smart contract is deployed, each organization that will participate in the auction is added to the endorsement policy.

Each round of the auction is stored as a separate entry in the blockchain ledger. As a result, all auction participants need to approve the creation of new rounds, the addition of bids and asks to the auction, or any other updates to the current state of the auction. As a result, none of the auction participants can unilaterally alter the auction outcome.

### Checks before creating and closing rounds

While organizations cannot alter the auction on the blockchain ledger, users could still use the dynamic nature of the auction to try to manipulate the auction outcome. For example, buyers and sellers could use the period while a round is active, while bids and asks are still being added to a round, to try to close the auction at a higher or lower price. As a result, the smart contract contains a series technical checks to ensure that the auction reaches an efficient price:

- Before a new auction is created, all organizations query their private data collection for the lowest ask from their organization. The open price of the auction (the reserve price) needs to be lower than the ask from any participating organization. This prevents a seller from starting the auction at an artificially high price.
- Before a new round is created, demand needs to be greater than supply for the previous round. Each organization also checks whether the previous round is still active by querying their private data collection to check that all asks from their organization with a lower price than the auction round has been added to the round. This prevents sellers from creating a new round and raising the price while sellers are still adding asks to the round and demand is temporarily greater than supply,
- Before a round can be closed, each organization checks that the supply of asks is greater than the demand from bids.
Organizations also query their private data collection for bids that are higher than the round price. This prevents buyers from closing the auction at an artificially low price, while buyers are still adding bids to the round and potentially pushing demand higher than supply. =

Because the checks are run by all participating organizations, any destructive updates to the auction can be rejected by organization, and prevents items from being sold at an artificially high or low price while participants are still interacting with the auction. This makes it possible for organizations to run the auction without coordination.

### Checks before adding bids and asks

Bidders and sellers may also have the incentive to manipulate their bids in response to the information that is revealed in the course of the auction. In the example below, bidder2 values the items being auctioned at 20. Bidder2 would not receive any consumer surplus if the auction ends in round 2. If the bidder lowers the quantity of their bid to 10, the auction would not progress to round 2 and the auction would close at a price of 15. Bidder3 would be better off if they changed their bid.

|  Round | Price | Bids:<br />Bidder - Quantity |Asks:<br />Seller - Quantity |
| -----------|-----------|---------|---------|
| 2 | 20 | bidder1 - 20<br />bidder2 - 10<br />bidder3 - 10  | seller1 - 20<br />seller2 - 20<br />seller3 - 20|
| 1 | 15 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 10 | seller1 - 20<br />seller2 - 20|
| 0 | 10 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20<br />bidder4 - 20 | seller1 - 20|

*Example 2: If bidder2 values the item at 20 each, they are better off altering their bid so the auction ends at round 1 instead of round 2.*

The smart contract uses a series of technical checks to prevent users from bidding strategically by changing their bid or asks in the course of the auction:

- All bidders need to join the first round in order to join subsequent rounds.
- All asks are automatically added to future rounds. This means that sellers cannot strategically withdraw asks in order to artificially lower supply.
- When a user creates a bid or ask, the price is kept private. However, a hash of the quantity of the bid or ask is hashed and stored on the public ledger. When a bid is revealed, each organization confirms that the hash of the bid that is revealed matches the hash on the ledger that was created before the auction. This ensures that users cannot change their bid during the auction.

## Economic mechanisms to prevent manipulation

In addition to being protected from malfeasance by the blockchain ledger and smart contract, the auction rules need to ensure that organizations have the incentive to agree to auction updates and can agree on the same market clearing price. The auction uses simple mechanism design to ensure that different organizations do not prefer that the auction is closed in different rounds and prices. As an example, the auction below contains sellers from two organizations, seller1 and seller2 from Org1 and seller3 from Org2. Seller1 and seller2 join the auction in earlier rounds, while seller3 joins in round 2. If the demand from bids were split equally, Org1 would sell 40 units for 15 in round 1, but only sell 28 at 20 in round 2. Org1 would prefer that the auction close at round 1, while Org2 would want the auction closed at round 2.

|  Round | Price | Bids:<br />Bidder - Quantity |Asks:<br />Seller - Org - Quantity |
| -----------|-----------|---------|---------|
| 2 | 20 | bidder1 - 20<br />bidder2 - 12<br />bidder3 - 10  | seller1 - Org1 - 20<br />seller2 - Org1 - 20<br />seller3 - Org2 - 20|
| 1 | 15 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20 | seller1 - Org1 - 20<br />seller2 - Org1 - 20|
| 0 | 10 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20<br />bidder4 - 20 | seller1 - Org1 - 20|

*Example 3: If bids were split equally among sellers, Org1 would prefer to close the auction at round 1.*

To prevent different organizations from disagreeing on the round price, the auction design introduces the concept of quantity **Sold**. When an ask is added to a round where demand is greater than supply, the seller is allocated the quantity of their ask. The seller then keeps this quantity for the next rounds. In the example below, seller1 sells 20 goods in round 1 and keeps that quantity for the remainder of the auction. The sum of all quantities assigned to each seller is the total quantity sold. A new seller can only sell a portion of their ask when demand is greater than the quantity sold. The auction is cleared when the quantity sold is equal to demand.

|  Round | Price | Bids:<br />Bidder - Quantity |Asks:<br />Seller - Org - Quantity | Quantity sold |
| -----------|-----------|---------|---------|--------|
| 2 | 40 | bidder1 - 20<br />bidder2 - 12<br />bidder3 - 10  | seller1 - Org1 - 20<br />seller2 - Org1 - 20<br />seller3 - Org2 - 20| 42 |
| 1 | 30 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20 | seller1 - Org1 - 20<br />seller2 - Org1 - 20| 40 |
| 0 | 20 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20<br />bidder4 - 20 | seller1 - Org1 - 20| 20 |

*Example 4: Seller1 and Seller2 .*

Revisiting the scenario in example 4, seller1 and seller2 still be awarded 20 each in the final round due to allocations in the previous rounds. Seller3 is allocated 2 in the final round because the demand of 42 is greater than the 40 sold by seller1 and seller2. Both organizations are better off if the auction was closed in round 2. To see the incentives created by this method it is worth seeing the auction from the perspective of the sellers.

|  Round | Seller1  | Seller2  Seller3 |
| -----------|-----------|---------|---------|
| 2 | quantity 20 - price 40 |  quantity 20 - price 40  | quantity 2 - price 40|
| 1 | quantity 20 - price 30 | quantity 20 - price 30 | |
| 0 | quantity 20 - price 20 | | |

*Example 5: Each seller is allocated a quantity that they can sell in an earlier round, after which their revenue can only increase as the price rises in future rounds.*

Each seller can only see their revenue increase as new rounds are created, as the price of the goods that they have already been allocated to sell increases. By separating the quantity that each seller sells from the price that they are awarded in the final round, each seller is incentivized to bid truthfully and try to sell as many items as possible in the earlier rounds. From the perspective each organization, every update to the auction that they endorse, whether creating new rounds or adding bids to the auction, increases the revenue to their sellers. The monotonicity of participant actions is what makes the auction incentive compatible.


## Get started: deploy the ascending auction smart contract

Now that we understand how the auction works, we can run an example auction on our local machine. We will use the Fabric test network to deploy the ascending auction smart contract.

If you have not already, you need to download the Fabric Docker images and the Fabric CLI tools. Make sure that you have installed all of the [Fabric prerequisites](https://hyperledger-fabric.readthedocs.io/en/latest/prereqs.html) and then follow the instructions to [Install the Fabric Samples, Binaries, and Docker Images](https://hyperledger-fabric.readthedocs.io/en/latest/install.html) in the Fabric documentation. In addition to downloading the Fabric images and tool binaries, the Fabric samples will also be cloned to your local machine.

Once you have cloned the fabric samples repo, you then need to checkout the branch with the auction smart contract. Navigate to the Fabric samples directory:
```
cd fabric-samples
```
Then run the following commands to checkout branch with the ascending auction sample:
```
git remote add auction https://github.com/nikhil550/fabric-samples.git
git fetch auction
git checkout --track auction/ascAuction
```

After you have checked out the branch, change into the test network directory.
```
cd fabric-samples/test-network
```

If the test network is already running, run the following command to bring the network down and start from a clean initial state.
```
./network.sh down
```

The Fabric test network contains two organizations, Org1 and Org2, that will run the auction. Each organization will deploy a Certificate Authority that will create the identities of bidders, sellers and auction administrator that will participate. Both organizations will run one peer that will store the bids and asks from their members. Run the following command to deploy a new network.  
```
./network.sh up createChannel -ca
```

Run the following command to deploy the auction smart contract to the network.
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

Before we can run the auction, we need to create the seller and bidder identities from Org1 and Org1, as well as the auction administrators that will interact with the auction on those users behalf. You can run the `auctionSetup.js` application to register the users that will participate in the auction. After the identities are created, each buyer and seller will create a bid or ask for tickets. The bids are only stored in the private data collection of their organization.
```
node auctionSetup.js
```

The `auctionSetup.js` program provides an example set of bids and asks that we will use to run the auction. However, you can run your own scenario by editing the program and updating the quantity and price preferred by each user, or change the number of identities participating in the auction.

### Start the auction

Instead of interacting the with network from a single program, Org1 and Org2 will run separate applications to demonstrate the non-cooperative nature of the auction. Each application will be run by the action administrator from the respective organization. The administrator will be able to read the bids submitted by the the organization members and submit those bids to an auction on those users behalf.

Open two new terminals, one for the Org1 application, and another for Org2. Open the Org1 terminal. Make sure that you are in the `ascending/application-javascript` directory:
```
cd fabric-samples/auction/ascending/application-javascript
```

Run the following command to start the Org1 application:
```
node appOrg1.js
```

The application connects to the network and starts a listener that waits for an auction to be started. You can see the application notify you when the listener has started:
```
<-- added contract listener
```

The listener will wait until an auction is started on the channel. When an auction has been created, the application will start a loop that contains the logic of how the auction administrator will interact with the auction. The program is designed to help each organization sell as many goods for the highest possible price:

- When a new round is created, each organization queries the their private data collection for the bids from buyers on the item sold by the auction. If organization finds that the bid is above the price of the round, the bid is added to the auction.
- When a new round is created, each organization also queries their private data collection for asks from seller. All Asks below the round price are added to the round.
- Each organization iterates through the rounds of the auction. When the quantity sold is less the the demand from bids in the final round, the organization tries to create a new round.
- If the quantity sold is the same as demand for the final round.

This logic will run until the listener learns that the final round has been closed, after which the auction is ended.

In your terminal for Org2, make sure that you are the same directory:

```
cd fabric-samples/auction/ascending/application-javascript
```

You can then start the Org2 application:
```
node appOrg2.js
```

The Org2 application starts the same listener and logic as Org1. However, Org2 application will also start an auction for tickets. After the auction is created, you can see each application start the logic of interacting with the auction using the auction administrator identity. You will see organization adding the bids and asks for tickets from their organization to the auction. You will also see the applications trying to create new rounds or close the final round. Many of those transactions will be rejected if the auction is still active. You will also see the occasional read/write error when the two applications try to update the auction at the same time. Neither application communicates or coordinates with each other.

The two test network organizations will eventually agree on a closing round and price for the auction. When the auction is closed and ended, you will see each organizations application print the final auction round:
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
cd ../../../test-network/
./network.sh down
````
