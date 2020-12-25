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

// Bid is used to add a users bid to the auction. The bid is stored in the private
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

	// create a composite key using the transaction ID
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{item,txID})
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

// SubmitBid is used by the bidder to add the hash of that bid stored in private data to the
// auction. Note that this function alters the auction in private state, and needs
// to meet the auction endorsement policy. Transaction ID is used identify the bid
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
		return err
	}

	// Check 1: the auction needs to be open for users to add their bid
	Status := auction.Status
	if Status != "open" {
		return fmt.Errorf("cannot join closed or ended auction")
	}

	// Check 2: the user needs to have joined the previous auction in order to
	// add their bid

	// create a composite key for bid using the transaction ID
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{auction.ItemSold, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	previousRound := round - 1

	if previousRound >= 0 {

		auctionLastRound, err := s.QueryAuctionRound(ctx, auctionID, previousRound)
		if err != nil {
			return err
		}

		previousBidders := make(map[string]Bidder)
		previousBidders = auctionLastRound.Bidders

		if _, previousBid := previousBidders[bidKey]; previousBid {

			//bid is in the previous auction, no action to take
		} else {
				return fmt.Errorf("bidder needs to have joined previous round")
		}
	}

	// check 3: check that bid has not changed on the public book

	publicBid, err := s.QueryPublic(ctx, auction.ItemSold, bidKeyType,txID)
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

	if !bytes.Equal(publicBid.Hash,bidHash) {
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
	auction.Demand = auction.Demand + NewBidder.Quantity
	bidders[bidKey] = NewBidder

	// quantity won will depend on whether supply is the same as demand
	if auction.Demand < auction.Quantity {
		for bidKey, bidder := range bidders {
			bidder.Won = bidder.Quantity
			bidders[bidKey] = bidder
		}
	} else if auction.Quantity == 0 {
		// quantity won is already zero
	}	else {
		for bidKey, bidder := range bidders {
			bidder.Won = (bidder.Quantity * auction.Demand) / auction.Quantity
			bidders[bidKey] = bidder
		}
	}
	auction.Bidders = bidders

	// create a composite for auction using the round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	newAuctionJSON, _ := json.Marshal(auction)

	err = ctx.GetStub().PutState(auctionKey, newAuctionJSON)
	if err != nil {
		return fmt.Errorf("failed to update auction: %v", err)
	}

	// create event for new bid
	err = ctx.GetStub().SetEvent("newBid", newAuctionJSON)
		if err != nil {
			return fmt.Errorf("event failed to register: %v", err)
		}

	return nil
}

// DeleteBid allows the submitter of the bid to delete their bid from the private data
// collection and from private state
func (s *SmartContract) DeleteBid(ctx contractapi.TransactionContextInterface, item string, txID string) (error) {

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

func (s *SmartContract) NewPublicBid(ctx contractapi.TransactionContextInterface, item string, txID string) error {

	//bidders cannot submit if there is an open auction
//	auction, err := s.QueryAuction(ctx, nil)
//	if err != nil {
//		return err
//	}

	// Find if round is closed. If a round is closed, declare found final.
//	for _, auctionRound := range auction {
//		if auctionRound.Status == "open" {
//			return fmt.Errorf("cannot add public bid or ask an auction for the item is open")
//		}
//  }

	// get the implicit collection name using the bidder's organization ID
	collection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

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
