/*
SPDX-License-Identifier: Apache-2.0
*/

package auction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"crypto/sha256"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const privateBidKeyType = "privateBid"
const publicBidKeyType = "publicBid"


// Bid is used to create a bid for a certain item. The bid is stored in the private
// data collection on the peer of the bidder's organization. The function returns
// the transaction ID so that users can identify and query their bid
func (s *SmartContract) Bid(ctx contractapi.TransactionContextInterface, item string) (string, error) {

	// get bid from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("error getting transient: %v", err)
	}

	privateBidJSON, ok := transientMap["privateBid"]
	if !ok {
		return "", fmt.Errorf("bid key not found in the transient map")
	}

	publicBidJSON, ok := transientMap["publicBid"]
	if !ok {
		return "", fmt.Errorf("bid key not found in the transient map")
	}

	// get the implicit collection name using the bidder's organization ID
	collection, err := getCollectionName(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// the bidder has to target their peer to store the bid
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return "", fmt.Errorf("Cannot store bid on this peer, not a member of this org: Error %v", err)
	}

	// the transaction ID is used as a unique index for the bid
	txID := ctx.GetStub().GetTxID()

	// create composite key for the private bid using the item and the txid
	privateBidKey, err := ctx.GetStub().CreateCompositeKey(privateBidKeyType, []string{item, txID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}

	// put the bid into the organization's implicit data collection
	err = ctx.GetStub().PutPrivateData(collection, privateBidKey, privateBidJSON)
	if err != nil {
		return "", fmt.Errorf("failed to input bid into collection: %v", err)
	}

	// create composite key for the public bid using the item and the txid
	publicBidKey, err := ctx.GetStub().CreateCompositeKey(publicBidKeyType, []string{item, txID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}

	// put the bid into the organization's implicit data collection
	err = ctx.GetStub().PutPrivateData(collection, publicBidKey, publicBidJSON)
	if err != nil {
		return "", fmt.Errorf("failed to input bid into collection: %v", err)
	}

	// return the trannsaction ID so that the uset can identify their bid
	return txID, nil
}

// SubmitBid adds a bid to an auction round. If successful, updates the
// quantity demanded and the quantity won by each bid
func (s *SmartContract) SubmitBid(ctx contractapi.TransactionContextInterface, auctionID string, round int, txID string) error {

	// get bid from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	transientBidJSON, ok := transientMap["publicBid"]
	if !ok {
		return fmt.Errorf("bid key not found in the transient map")
	}

	auction, err := s.QueryAuctionRound(ctx, auctionID, round)
	if err != nil {
		return fmt.Errorf("Error getting auction round from state")
	}

	// create a composite key for bid using the transaction ID
	publicBidKey, err := ctx.GetStub().CreateCompositeKey(publicBidKeyType, []string{auction.ItemSold, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// Check 1: the auction needs to be open for users to add their bid
	status := auction.Status
	if status != "open" {
		return fmt.Errorf("cannot join closed or ended auction")
	}

	// Check 2: the user needs to have joined the previous auction in order to
	// add their bid
	previousRound := round - 1

	if previousRound >= 0 {

		auctionLastRound, err := s.QueryAuctionRound(ctx, auctionID, previousRound)
		if err != nil {
			return fmt.Errorf("cannot pull previous auction round from state")
		}

		previousBidders := make(map[string]Bidder)
		previousBidders = auctionLastRound.Bidders

		if _, previousBid := previousBidders[publicBidKey]; previousBid {

			//bid is in the previous auction, no action to take
		} else {
			return fmt.Errorf("bidder needs to have joined previous round")
		}
	}

	// check 3: check that bid has not changed on the public book
	publicBid, err := s.QueryPublic(ctx, auction.ItemSold, publicBidKeyType, txID)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from public order book: %v", err)
	}

	collection := "_implicit_org_" + publicBid.Org

	bidHash, err := ctx.GetStub().GetPrivateDataHash(collection, publicBidKey)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from collection: %v", err)
	}
	if bidHash == nil {
		return fmt.Errorf("bid hash does not exist: %s", bidHash)
	}

	hash := sha256.New()
	hash.Write(transientBidJSON)
	calculatedBidJSONHash := hash.Sum(nil)

	// verify that the hash of the passed immutable properties matches the on-chain hash
	if !bytes.Equal(calculatedBidJSONHash, bidHash) {
		return fmt.Errorf("hash %x for bid JSON %s does not match hash in auction: %x",
			calculatedBidJSONHash,
			transientBidJSON,
			bidHash,
		)
	}

	if !bytes.Equal(publicBid.Hash, bidHash) {
		return fmt.Errorf("Bidder has changed their bid")
	}

	var bid *PublicBid
	err = json.Unmarshal(transientBidJSON, &bid)
	if err != nil {
		return err
	}

	// now that all checks have passed, create new bid
	newBidder := Bidder{
		Buyer:    bid.Buyer,
		Org:      bid.Org,
		Quantity: bid.Quantity,
		Won:      0,
	}

	// add the bid to the new list of bidders
	bidders := make(map[string]Bidder)
	bidders = auction.Bidders
	bidders[publicBidKey] = newBidder

	newDemand := 0
	for _, bidder := range bidders {
		newDemand = newDemand + bidder.Quantity
	}
	auction.Demand = newDemand

	// If quantity = sold, no need to update the asks, just allocate sold amount
	// to smarller bids first
	if (auction.Quantity == auction.Sold) && (auction.Sold != 0) {

		if auction.Quantity >= auction.Demand {

			for bid, bidder := range bidders {
				bidder.Won = bidder.Quantity
				bidders[bid] = bidder
			}
		} else {

			for bid, bidder := range bidders {
				bidder.Won = (bidder.Quantity * auction.Sold) / auction.Demand
				bidders[bid] = bidder
			}
		}
		auction.Bidders = bidders
	} else if auction.Sold == 0 {
		auction.Bidders = bidders
	} else {
		sellers := make(map[string]Seller)
		sellers = auction.Sellers

		previousSold := auction.Sold

		newSold := 0
		if auction.Quantity >= auction.Demand {
			newSold = auction.Demand
			remainingSold := newSold - previousSold
			for bid, bidder := range bidders {
				bidder.Won = bidder.Quantity
				bidders[bid] = bidder
			}
			totalUnsold := 0
			for _, seller := range sellers {
				totalUnsold = totalUnsold + seller.Unsold
			}
			for ask, seller := range sellers {
				seller.Sold = seller.Sold + (seller.Unsold*remainingSold)/totalUnsold
				seller.Unsold = seller.Quantity - seller.Sold
				sellers[ask] = seller
			}
		} else {
			newSold = auction.Quantity
			for bid, bidder := range bidders {
				bidder.Won = (bidder.Quantity * auction.Sold) / auction.Demand
				bidders[bid] = bidder
			}
			for ask, seller := range sellers {
				seller.Sold = seller.Quantity
				seller.Unsold = 0
				sellers[ask] = seller
			}
		}
		auction.Sold = newSold

		auction.Bidders = bidders
		auction.Sellers = sellers
	}

	// create a composite for the auction round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	newAuctionJSON, _ := json.Marshal(auction)

	// put the updated auction round in state
	err = ctx.GetStub().PutState(auctionKey, newAuctionJSON)
	if err != nil {
		return fmt.Errorf("failed to update auction: %v", err)
	}

	return nil
}

// DeleteBid allows the submitter of the bid to delete their bid from the private data
// collection and from private state
func (s *SmartContract) DeleteBid(ctx contractapi.TransactionContextInterface, item string, txID string) error {

	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	collection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	publicBidKey, err := ctx.GetStub().CreateCompositeKey(publicBidKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().DelPrivateData(collection, publicBidKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", publicBidKey, err)
	}

	privateBidKey, err := ctx.GetStub().CreateCompositeKey(privateBidKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = s.checkBidOwner(ctx, collection, privateBidKey)
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelPrivateData(collection, privateBidKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", privateBidKey, err)
	}

	return nil
}

// NewPublicBid adds a bid to the public order book. This ensures
// that bidders cannot change their bid during an active auction
func (s *SmartContract) NewPublicBid(ctx contractapi.TransactionContextInterface, item string, txID string) error {

	// get the implicit collection name using the bidder's organization ID
	collection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// create composite key for the bid using the item and the txid
	publicBidKey, err := ctx.GetStub().CreateCompositeKey(publicBidKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	hash, err := ctx.GetStub().GetPrivateDataHash(collection, publicBidKey)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from collection: %v", err)
	}
	if hash == nil {
		return fmt.Errorf("bid hash does not exist: %s", publicBidKey)
	}

	// get the org of the subitting bidder
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	// store the hash along with the seller's organization in the public order book
	publicBid := BidAskHash{
		Org:  clientOrgID,
		Hash: hash,
	}

	publicBidJSON, _ := json.Marshal(publicBid)

	// put the ask hash of the bid in the public order book
	err = ctx.GetStub().PutState(publicBidKey, publicBidJSON)
	if err != nil {
		return fmt.Errorf("failed to input bid into state: %v", err)
	}

	return nil
}
