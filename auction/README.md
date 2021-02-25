## Auction sample

The auction sample provides a set of smart contracts and tutorials that implement a series of distributed auctions. Each example uses important Fabric features such as private data, access control, and state-based endorsement.

- [Simple blind auction](simple-blind-auction): This example implements an auction where a single good is sold to the highest bidder. Because the tutorial provides more details on how the auction is implemented, users should go through this example first.
- [Dutch auction](dutch-auction): This example implements an auction in which multiple items of the same type can be sold to more than one buyer. This example also includes an auditor that can update the auction in the case that auction participants disagree. The smart contracts provides an example of how to use a more complicated signature policy to govern an asset using state based endorsement.
- [Ascending auction](ascending-auction): This example demonstrates how smart contracts can can be used to run a fully decentralized auction. Organizations that act in their self interest run an auction that reaches a market clearing price without relying on a central auctioneer.
