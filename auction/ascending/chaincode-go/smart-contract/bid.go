/*
SPDX-License-Identifier: Apache-2.0
*/

package auction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const bidKeyType = "bid"

// Bid is used to create a bid for a certain item. The bid is stored in the private
// data collection on the peer of the bidder's organization. The function returns
// the transaction ID so that users can identify and query their bid
func (s *SmartContract) Bid(ctx contractapi.TransactionContextInterface, item string) (string, error) {

	// get bid from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("error getting transient: %v", err)
	}

	BidJSON, ok := transientMap["bid"]
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

	// create composite key for the bid using the item and the txid
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{item, txID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}

	// put the bid into the organization's implicit data collection
	err = ctx.GetStub().PutPrivateData(collection, bidKey, BidJSON)
	if err != nil {
		return "", fmt.Errorf("failed to input bid into collection: %v", err)
	}

	// return the trannsaction ID so that the uset can identify their bid
	return txID, nil
}

// SubmitBid adds a bid to an auction round. If successful, updates the
// quantity demanded and the quantity won by each bid
func (s *SmartContract) SubmitBid(ctx contractapi.TransactionContextInterface, auctionID string, round int, quantity int, txID string) error {

	// get identity of submitting client
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	// get the org of the subitting bidder
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	auction, err := s.QueryAuctionRound(ctx, auctionID, round)
	if err != nil {
		return fmt.Errorf("Error getting auction round from state")
	}

	// Check 1: the auction needs to be open for users to add their bid
	Status := auction.Status
	if Status != "open" {
		return fmt.Errorf("cannot join closed or ended auction")
	}

	// create a composite key for bid using the transaction ID
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{auction.ItemSold, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	previousRound := round - 1

	if previousRound >= 0 {

		auctionLastRound, err := s.QueryAuctionRound(ctx, auctionID, previousRound)
		if err != nil {
			return fmt.Errorf("cannot pull previous auction round from state")
		}

		previousBidders := make(map[string]Bidder)
		previousBidders = auctionLastRound.Bidders

		// Check 2: the user needs to have joined the previous auction in order to
		// add their bid
		if _, previousBid := previousBidders[bidKey]; previousBid {

			//bid is in the previous auction, no action to take
		} else {
			return fmt.Errorf("bidder needs to have joined previous round")
		}
	}

	// check 3: check that bid has not changed on the public book
	publicBid, err := s.QueryPublic(ctx, auction.ItemSold, bidKeyType, txID)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from public order book: %v", err)
	}

	collection := "_implicit_org_" + publicBid.Org

	bidHash, err := ctx.GetStub().GetPrivateDataHash(collection, bidKey)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from collection: %v", err)
	}
	if bidHash == nil {
		return fmt.Errorf("bid hash does not exist: %s", bidHash)
	}

	if !bytes.Equal(publicBid.Hash, bidHash) {
		return fmt.Errorf("Bidder has changed their bid")
	}

	// now that all checks have passed, create new bid
	NewBidder := Bidder{
		Buyer:    clientID,
		Org:      clientOrgID,
		Quantity: quantity,
		Won:      0,
	}

	// add the bid to the new list of bidders
	bidders := make(map[string]Bidder)
	bidders = auction.Bidders
	bidders[bidKey] = NewBidder

	newDemand := 0
	for _, bidder := range bidders {
		newDemand = newDemand + bidder.Quantity
	}

	auction.Demand = newDemand
	// quantity won will depend on whether supply is the same as demand
	if auction.Demand < auction.Quantity {
		for bid, bidder := range bidders {
			bidder.Won = bidder.Quantity
			bidders[bid] = bidder
		}
	} else if auction.Quantity == 0 {
		// quantity won is already zero
	} else {
		for bid, bidder := range bidders {
			bidder.Won = (bidder.Quantity * auction.Quantity) / auction.Demand
			bidders[bid] = bidder
		}
	}
	auction.Bidders = bidders

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

	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = s.checkBidOwner(ctx, collection, bidKey)
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelPrivateData(collection, bidKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", bidKey, err)
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
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	Hash, err := ctx.GetStub().GetPrivateDataHash(collection, bidKey)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from collection: %v", err)
	}
	if Hash == nil {
		return fmt.Errorf("bid hash does not exist: %s", bidKey)
	}

	// get the org of the subitting bidder
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	// store the hash along with the seller's organization in the public order book
	publicBid := BidAskHash{
		Org:  clientOrgID,
		Hash: Hash,
	}

	publicBidJSON, _ := json.Marshal(publicBid)

	// put the ask hash of the bid in the public order book
	err = ctx.GetStub().PutState(bidKey, publicBidJSON)
	if err != nil {
		return fmt.Errorf("failed to input bid into state: %v", err)
	}

	return nil
}
