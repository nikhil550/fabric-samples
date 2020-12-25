/*
SPDX-License-Identifier: Apache-2.0
*/

package auction

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const askKeyType = "ask"

func (s *SmartContract) Ask(ctx contractapi.TransactionContextInterface, item string) (string, error) {

	// get bid from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("error getting transient: %v", err)
	}

	askJSON, ok := transientMap["ask"]
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
	askKey, err := ctx.GetStub().CreateCompositeKey(askKeyType, []string{item,txID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}

	// put the bid into the organization's implicit data collection
	err = ctx.GetStub().PutPrivateData(collection, askKey, askJSON)
	if err != nil {
		return "", fmt.Errorf("failed to input price into collection: %v", err)
	}

	// return the trannsaction ID so that the uset can identify their bid
	return txID, nil
}

func (s *SmartContract) SubmitAsk(ctx contractapi.TransactionContextInterface, auctionID string, round int, quantity int, txID string) error {

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

	// the auction needs to be open for users to add their bid
	Status := auction.Status
	if Status != "open" {
		return fmt.Errorf("cannot join closed or ended auction")
	}

	// store the hash along with the bidder's organization
	NewSeller := Seller{
		Seller:   clientID,
		Org:      clientOrgID,
		Quantity: quantity,
	}

	// create a composite key for ask using the transaction ID
	askKey, err := ctx.GetStub().CreateCompositeKey(askKeyType, []string{auction.ItemSold, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// add to the list of sellers
	sellers := make(map[string]Seller)
	sellers = auction.Sellers
	sellers[askKey] = NewSeller
	auction.Sellers = sellers

	auction.Quantity = auction.Quantity + NewSeller.Quantity

	// Update the list of winners
	bidders := make(map[string]Bidder)
	bidders = auction.Bidders

	if auction.Demand < auction.Quantity {
		for bidKey, bidder := range bidders {
			bidder.Won = bidder.Quantity
			bidders[bidKey] = bidder
		}
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

	// Create event for new ask
	err = ctx.GetStub().SetEvent("newAsk", newAuctionJSON)
		if err != nil {
			return fmt.Errorf("event failed to register: %v", err)
		}

	return nil
}

// DeleteAsk allows the seller of the bid to delete their bid from the private data
// collection and from private state
func (s *SmartContract) DeleteAsk(ctx contractapi.TransactionContextInterface, item string, txID string) error {

	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	collection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	askKey, err := ctx.GetStub().CreateCompositeKey(askKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = s.checkAskOwner(ctx, collection, askKey)
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelPrivateData(collection, askKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", askKey, err)
	}

	return nil
}


func (s *SmartContract) NewPublicAsk(ctx contractapi.TransactionContextInterface, item string, txID string) error {

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

	askKey, err := ctx.GetStub().CreateCompositeKey(askKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	Hash, err := ctx.GetStub().GetPrivateDataHash(collection, askKey)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from collection: %v", err)
	}
	if Hash == nil {
		return fmt.Errorf("bid hash does not exist: %s", askKey)
	}

	// get the org of the subitting bidder
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	// store the hash along with the seller's organization in the public order book
	publicAsk := BidAskHash{
		Org:  clientOrgID,
		Hash: Hash,
	}

	publicAskJSON, _ := json.Marshal(publicAsk)

	// put the ask hash of the bid in the public order book
	err = ctx.GetStub().PutState(askKey, publicAskJSON)
	if err != nil {
		return fmt.Errorf("failed to input bid into state: %v", err)
	}

	return nil
}
