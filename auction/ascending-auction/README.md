## Ascending auction

This example uses a Hyperledger Fabric smart contract to run a decentralized auction. Instead of being run by a central auctioneer, the auction is run by self interested organizations who are members of a Fabric network. Each organization has members that are buyers and sellers of a good being sold using an auction. When running the auction, each organization acts on behalf of its sellers and attempts to fill as many asks from its members at the highest possible price. The smart contract contains technical and economic mechanisms that prevent organizations from manipulating the outcome to their benefit and sets rules that incentivize organizations to agree to the items being sold at an efficient market clearing price. As a result, the members of the network can participate in a competitive auction without the presence of a central intermediary.

The smart contract implements an ascending price auction that can sell multiple items of the same type. The auction can be used by multiple sellers, as long as they are providing the same good. Buyers and sellers create bids and asks for the item before the auction starts and submit their bids and asks to the auction after it is created. The price of the auction rises until supply is sufficient to meet demand. Although the auction can be used for many situations, it is designed to allocate goods quickly against pre-existing bids and asks, such as energy entering the electricity grid or goods moving through a supply chain.

This tutorial discusses the design of the auction and how the blockchain network protects the auction from manipulation by self interested users. You can then use this tutorial to run an example auction on a running Fabric network. To demonstrate the decentralized nature of the auction, the example is run by submitting bids and asks using two applications from different organizations. The applications interact with the auction without any coordination or cooperation. However, the organizations running the applications and operating the Fabric network are able to use the smart contract to close the auction at an efficient price.

- [Auction design](#auction-design)
- [Technical mechanisms to prevent manipulation](#technical-mechanisms-to-prevent-manipulation)
- [Economic mechanisms](#economic-mechanisms-to-prevent-manipulation)
- [Get started: deploy the ascending auction smart contract](#get-started-deploy-the-ascending-auction-smart-contract)
- [Run the auction as Org1 and Org2](#get-started-deploy-the-ascending-auction-smart-contract)

## Auction design

The auction is run for individual buyers and sellers that are members of the organizations that run the Fabric network. Buyers and sellers create bids or asks to buy or sell a homogenous good. Each bid (or ask) specifies the quantity of the good that the user is willing to buy (or sell) at a given price. Only the buyer or seller of the good can create or remove the bid. Bids and asks are stored in the private data collection of the users organization. As a result, the price at which users are willing to buy or sell is not shared with other organizations in the network.

Each auction consists of multiple rounds. Each round sets a price at which good can be bought or sold. Users add their bid or ask to each round separately, announcing how much of the they are willing to buy or sell at the round price. When the quantity demanded by bids is greater than the supply provided by asks, a new round of the auction is created. The new round raises the price by a set increment and allows buyers and sellers to submit new bids. Users only reveal the quantity that they are willing to buy or sell at a given price and keep the rest of their bid private. When the quantity supplied in a round is sufficient to meet demand, the auction is closed and the price of the good is set as the price of the final round.

|  Round | Price | Bids:<br />Bidder - Quantity |Asks:<br />Seller - Quantity |
| -----------|-----------|---------|---------|
| 2 | 20 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20  | seller1 - 20<br />seller2 - 20<br />seller3 - 20|
| 1 | 15 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20 | seller1 - 20<br />seller2 - 20|
| 0 | 10 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20<br />bidder4 - 20 | seller1 - 20|

*Example 1: Each auction consists of multiple rounds. The price of each round is raised by a set increment when a new round is created. Bids and asks are added separately to each round.*

Auctions can be started by any user who is a member of the network. Bids or asks can be added to auction rounds either by the bid creator or by an auction administrator from the users organization. Auction administrators make it easier for buyers and sellers to participate, since they only have to create the bid or ask once and then have the administrator submit the bid to the rounds of the auction on their behalf. Auction administrators are organization members that are identified by [attribute based access control](https://hyperledger-fabric-ca.readthedocs.io/en/latest/users-guide.html#attribute-based-access-control). In addition to adding their bid or ask, any user can try to create a new round and raise the auction price. Network members can also try to close the auction to set the final price. Any time an auction is created, the auction becomes a non-cooperative game between buyers and sellers. Sellers can try to create new rounds to raise the auction price while buyers try to close the auction and purchase goods for the lowest possible price. Auction administrators will try to fill as many asks from sellers as they can, while also adding bids to the relevant rounds to increase demand.

## Technical mechanisms to prevent manipulation

Because the ascending auction smart contract is meant to be run without a central auctioneer, the self interested organizations that run the auction need to be prevented from altering the auction outcome. The smart contract and the Fabric blockchain work together to prevent organizations from manipulating the auction.

### Smart contract endorsement policy

Assets that are stored on a Fabric blockchain ledger are protected by the smart contract endorsement policy. A sufficient number of organizations need to agree before an update can be made to the blockchain ledger. When the ascending auction smart contract is deployed, all the organizations that will participate in the auction are added to the endorsement policy.

Each round of the auction is stored as a separate entry in the blockchain ledger. As a result, all auction participants need to approve any updates to existing rounds or the creation of new rounds. As a result, none of the auction participants can unilaterally alter the bids submitted to the auction or change the auction price.

### Checks for creating and closing rounds

While organizations cannot alter the auction on the blockchain ledger, users could still use the dynamic nature of the auction to try to manipulate the auction outcome. For example, buyers and sellers could use the period when a round is still active, while bids and asks are still being added, to try to close the auction at an artificially high or lower price. As a result, each organization participating in the auction runs a series of checks to prevent other organizations from submitting harmful updates:

- Before a new auction is created, all organizations search for the lowest ask from their sellers in their private data collection. Organizations will reject any opening price that is higher than this lowest ask. This prevents sellers from trying to start the auction at an artificially high price.
- Before a new round is created, organizations verify that the demand from asks is greater than the supply from bids in the previous round. Organizations also check that all the asks lower than the previous round price from their organization have been added to the previous round before the new round can be created. This prevents sellers from raising the auction price while sellers are still adding asks and demand may only be temporarily greater than supply.
- Before the auction can be closed, organizations verify that supply is sufficient to meet demand. Organizations also verify that all bidders from their organization with a higher price have been added to the final round. This prevents buyers from closing the auction at an artificially low price, while buyers are still adding bids that can increase demand.

Because all organizations with buyers and sellers participating need to approve auction updates, no individual user can intentionally or accidentally sell the items at an inefficient price. This allows organizations to run the auction without to coordinate with each other.

### Checks for adding bids and asks

Bidders and sellers may also try to manipulate their bids in response to the information being revealed in the course of the auction. In the example below, bidder2 values the items being auctioned at 20. If the auction ends in round 2, bidder2 would not receive any consumer surplus. However, if bidder2 lowered the quantity of their bid to 10 in round 1, the auction would close at a price of 15 and bidder2 would be better off.

|  Round | Price | Bids:<br />Bidder - Quantity |Asks:<br />Seller - Quantity |
| -----------|-----------|---------|---------|
| 2 | 20 | bidder1 - 20<br />bidder2 - 10<br />bidder3 - 10  | seller1 - 20<br />seller2 - 20<br />seller3 - 20|
| 1 | 15 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 10 | seller1 - 20<br />seller2 - 20|
| 0 | 10 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20<br />bidder4 - 20 | seller1 - 20|

*Example 2: If bidder2 values the item at 20 each, they are better off altering their bid so the auction ends at round 1 instead of round 2.*

To prevent buyers or sellers from acting strategically, each organization runs a series of checks before a bid or ask can be added to a round:

- All bidders need to join the first round in order to join subsequent rounds.
- All asks are automatically added to future rounds. This means that sellers cannot strategically withdraw asks in order to artificially limit supply.
- When a user creates a bid or ask, a hash of the quantity is created and stored on the public ledger. When the bid or ask is added to a round, each organization confirms that the hash of the bid quantity matches the hash that was created before the auction. This ensures that users cannot change their bid during the auction.

## Economic mechanisms

In addition to being protected from manipulation by the blockchain network, the rules of the auction need to ensure that the participating organizations can agree on the same market clearing price and have the incentive to agree to auction updates. The auction uses simple mechanism design prevent organizations from preferring outcomes. As an example, the auction below contains sellers from two organizations; Seller1 and seller2 are from Org1 while seller3 is from Org2. Seller1 and seller2 join the auction in earlier rounds, while seller3 joins in round 2. If the demand from bids were split equally, Org1 would sell 40 items for 15 in round 1, but only sell 28 items for 20 in round 2. Assuming that they are acting on behalf of their sellers, Org1 would prefer that the auction close at round 1, while Org2 would be better off if the auction closed at round 2.

|  Round | Price | Bids:<br />Bidder - Quantity |Asks:<br />Seller - Org - Quantity |
| -----------|-----------|---------|---------|
| 2 | 20 | bidder1 - 20<br />bidder2 - 12<br />bidder3 - 10  | seller1 - Org1 - 20<br />seller2 - Org1 - 20<br />seller3 - Org2 - 20|
| 1 | 15 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20 | seller1 - Org1 - 20<br />seller2 - Org1 - 20|
| 0 | 10 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20<br />bidder4 - 20 | seller1 - Org1 - 20|

*Example 3: If bids were split equally among sellers, Org1 would prefer to close the auction at round 1.*

To prevent different organizations from disagreeing on price, the auction design introduces the concept of quantity **sold**. When an ask is added to a round where demand is greater than supply, the seller is allocated the quantity of their ask. The seller then keeps these items in additional rounds. In the example below, seller1 sells 20 goods in round 1 and keeps that quantity for the remainder of the auction. The sum of items allocated to each seller is the quantity that is sold for the round. A new ask can only be filled, either partially or in full, when demand is greater than what was allocated in previous rounds. The auction can be closed when the quantity sold is equal to demand.

|  Round | Price | Bids:<br />Bidder - Quantity |Asks:<br />Seller - Org - Quantity | Quantity sold |
| -----------|-----------|---------|---------|--------|
| 2 | 40 | bidder1 - 20<br />bidder2 - 12<br />bidder3 - 10  | seller1 - Org1 - 20<br />seller2 - Org1 - 20<br />seller3 - Org2 - 20| 42 |
| 1 | 30 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20 | seller1 - Org1 - 20<br />seller2 - Org1 - 20| 40 |
| 0 | 20 | bidder1 - 20<br />bidder2 - 20<br />bidder3 - 20<br />bidder4 - 20 | seller1 - Org1 - 20| 20 |

*Example 4: Seller1 is allocated 20 items in round 0, while seller2 is allocated 20 items in round 1. In round 2, 40 items have already been sold. As a result, seller3 is allocated the remaining demand.*

Revisiting the scenario in example 4, both organizations prefer that the auction is closed in round 2. What happens if Org1 tried to prevent seller3 from joining the auction in the hope of trying to create a new round? The quantity sold is only allocated when a new round is created or when the auction is closed, after all users have added their bids and asks to the auction. As a result, the quantity sold or price received by each organization can only increase when all organizations agree. To see the incentives created by this design, it is worth seeing the auction from the perspective of the sellers.

|  Round | Seller1  | Seller2 | Seller3 |
| -----------|-----------|---------|---------|
| 2 | quantity 20 - price 40 |  quantity 20 - price 40  | quantity 2 - price 40|
| 1 | quantity 20 - price 30 | quantity 20 - price 30 | |
| 0 | quantity 20 - price 20 | | |

*Example 5: Each seller is allocated items earlier in the auction, after which their revenue can only increase as the price rises in future rounds.*

Each seller can only see their revenue increase when new rounds are created, raising the price of the goods that they have already sold. Because the quantity allocated to each seller is independent of the price that they receive, each seller is best off when they bid truthfully and try sell as many items as possible in earlier rounds. From the perspective of each organization, every update to the auction that they can endorse, such as adding new bids and asks to the auction, creating new rounds, or closing the auction, increases the revenue earned by their sellers. The auction is incentive compatible because every action taken by the organization running the auction has a monotone effect and increases the revenue of their sellers.

## Get started: deploy the ascending auction smart contract

Now that we understand how the auction works, we can run an example auction on our local machine. We will use the Fabric test network to deploy the ascending auction smart contract.

If you have not already, you need to download the Fabric Docker images and the Fabric CLI tools. Make sure that you have installed all of the [Fabric prerequisites](https://hyperledger-fabric.readthedocs.io/en/latest/prereqs.html) and then follow the instructions to [Install the Fabric Samples, Binaries, and Docker Images](https://hyperledger-fabric.readthedocs.io/en/latest/install.html) in the Fabric documentation. In addition to downloading the Fabric images and tool binaries, the Fabric samples will also be cloned to your local machine.

Once you have cloned the fabric samples repo, you then need to checkout the branch with the auction smart contract. Navigate to the Fabric samples directory:
```
cd fabric-samples
```
Then run the following commands to checkout the branch with the ascending auction sample:
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

The Fabric test network contains two organizations, Org1 and Org2, that will run the auction smart contract. Each organization will deploy a Certificate Authority that will create the identities of bidders, sellers and auction administrator that will participate in the auction. Both organizations will run one peer that will store the bids and asks of their members. Run the following command to deploy a new network and create a channel that will become the auction platform:
```
./network.sh up createChannel -ca
```

When the network is up, run the following command to deploy the ascending auction smart contract to the channel:
```
./network.sh deployCC -ccn auction -ccp ../auction/ascending-auction/chaincode-go/ -ccl go -ccep "AND('Org1MSP.peer','Org2MSP.peer')"
```

## Run the auction as Org1 and Org2

After you deployed the auction smart contract, we can run the auction using a set of Node.js applications. Change into the `application-javascript` directory:
```
cd fabric-samples/auction/ascending-auction/application-javascript
```

From this directory, run the following command to download the application dependencies:
```
npm install
```

### Register users and submit bids and asks

Before we can run the auction, we need to create the seller and bidder identities from Org1 and Org2, as well as the auction administrators that will interact with the auction on those users behalf. You can run the `auctionSetup.js` application to register the organization members that will participate in the auction. After the user identities are registered and enrolled with their organizations Certificate Authority, each buyer and seller will create a bid or ask for tickets to an event.
```
node auctionSetup.js
```

The `auctionSetup.js` program provides an example set of bids and asks that we will use to run the auction. However, you can create your own auction scenario editing the example bids or creating additional users.

### Start the auction

Instead of interacting the network with a single application, Org1 and Org2 will run separate programs to demonstrate the competitive nature of the auction. The application of each organization will be run by their application administrator. The auction administrators are able to read the bid and asks from their organization and can submit those bids and asks to an active auction.

Open two new command terminals. We will use one terminal to run the Org1 application and another terminal for Org2. Open the terminal that you will use for Org1 and make sure that you are in the `ascending/application-javascript` directory:
```
cd fabric-samples/auction/ascending-auction/application-javascript
```

Run the following command to start the Org1 application:
```
node appOrg1.js
```

The application connects to the channel and adds a listener to the ascending auction smart contract. You can see the application notify you that the listener has started:
```
<-- added contract listener
```

The listener will notify the application when a new auction has been created. The auction administrator will then query the auction on blockchain ledger to learn the item that the auction is selling. After the auction has started, the auction administrator follows a logic that allows the organization to fill as many asks from its sellers for the highest possible price:

- When a new round is created, the application reads the bids from buyers for the item being sold by the auction. Any bids that are above the round price are added to the auction round.
- When a new round is created, the application also reads the asks for the item from the organizations sellers. Any asks that are below the round price are added to the auction round.
- When the quantity sold is less than demand in the round with the highest price, the application will try to create a new round.
- If the quantity sold is the same as demand in the round with the highest price, the application will try to close the auction.

The application will run the logic above until the listener learns that the auction has been closed. After the auction closed, the auction is formally ended, and all rounds but the final round will be removed.

Now that Org1 is waiting for an auction to be created, we can create the auction using the Org2 application. Open the terminal for Org2, make sure that you are the same directory:

```
cd fabric-samples/auction/ascending/application-javascript
```

You can then start the Org2 application:
```
node appOrg2.js
```

The Org2 application starts the same smart contract listener as Org1 and uses the same logic to maximize revenue from the auction. However, Org2 application will also start an auction for tickets. After the auction is created, the smart contract will emit an event that informs both Org1 and Org2 that the auction has been created. Each application will start adding bids and asks to the auction. The applications will also try to raise the auction price by creating new rounds or close the round with the highest price. Many of those transactions will be rejected if the other organization is still adding asks or bids to the auction round. You will also see the occasional read/write error when the two applications try to update the auction at the same time. Neither application communicates with each other.

While the two applications in our example auction are run by organization administrators, it is possible for buyers and sellers to interact with the auction directly. Bidders would add their bids and try to close the auction at the lowest possible price, while sellers would add their asks and try to raise the auction price by creating new rounds. The same checks that allow the two auction administrators to interact with the auction without coordinating would allow buyers and sellers to follow their self interest without changing the auction outcome.

The two test network organizations will eventually agree on a closing price for the auction. When the auction is closed and ended, you will see each organizations application print the final auction round:
```
*** Result: Final auction round: [
  {
    "objectType": "auction",
    "id": "auction1",
    "round": 5,
    "status": "closed",
    "item": "tickets",
    "price": 45,
    "quantity": 35,
    "sold": 30,
    "demand": 30,
    "sellers": {
      "\u0000publicAsk\u0000tickets\u00005fc2653ce02672aafc4dfb55f50bf4909dc9c4a94284ed9b90e1d2c46871fb2f\u0000": {
        "seller": "x509::CN=seller2,OU=client+OU=org1+OU=department1::CN=ca.org1.example.com,O=org1.example.com,L=Durham,ST=North Carolina,C=US",
        "org": "Org1MSP",
        "quantity": 10,
        "sold": 10,
        "unsold": 0
      },
      "\u0000publicAsk\u0000tickets\u000088a7664f22e24df09cfeccf7644728b78c33e3dcc54473c5b71bad0a6015cec5\u0000": {
        "seller": "x509::CN=seller3,OU=client+OU=org2+OU=department1::CN=ca.org2.example.com,O=org2.example.com,L=Hursley,ST=Hampshire,C=UK",
        "org": "Org2MSP",
        "quantity": 10,
        "sold": 5,
        "unsold": 5
      },
      "\u0000publicAsk\u0000tickets\u00009aabf35b36abd5c3b7017da759043676b3c1120d69cc67d1c63087240c6d68e6\u0000": {
        "seller": "x509::CN=seller1,OU=client+OU=org1+OU=department1::CN=ca.org1.example.com,O=org1.example.com,L=Durham,ST=North Carolina,C=US",
        "org": "Org1MSP",
        "quantity": 15,
        "sold": 15,
        "unsold": 0
      }
    },
    "bidders": {
      "\u0000publicBid\u0000tickets\u0000327a0cd2efc7b00c2396a774ab7ea904bd73d25c4693c839e7a3a2b54bc670e1\u0000": {
        "buyer": "x509::CN=bidder5,OU=client+OU=org2+OU=department1::CN=ca.org2.example.com,O=org2.example.com,L=Hursley,ST=Hampshire,C=UK",
        "org": "Org2MSP",
        "quantityBid": 15,
        "quantityWon": 15
      },
      "\u0000publicBid\u0000tickets\u0000b58e29b011c9d48b775f636375cc2c0658f305738d3e5e9bfcf9ea0690effb29\u0000": {
        "buyer": "x509::CN=bidder4,OU=client+OU=org2+OU=department1::CN=ca.org2.example.com,O=org2.example.com,L=Hursley,ST=Hampshire,C=UK",
        "org": "Org2MSP",
        "quantityBid": 15,
        "quantityWon": 15
      }
    }
  }
]
```

## Clean up

When you are done using the auction smart contract, you can bring down the network and clean up the environment. In the `application-javascript` directory, run the following command to remove the wallets used to run the applications:
```
rm -rf wallet
```

You can then navigate to the test network directory and bring down the network:
````
cd ../../../test-network/
./network.sh down
````
