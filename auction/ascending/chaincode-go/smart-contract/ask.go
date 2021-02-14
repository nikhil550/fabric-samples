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

// Ask is used to sell a certain item. The ask is stored in private data
// of the sellers organization, and identified by the item and transaction id
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

	// create a composite key using the item and transaction ID
	askKey, err := ctx.GetStub().CreateCompositeKey(askKeyType, []string{item, txID})
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

// SubmitAsk is used to add an ask to an active auction round
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
		return fmt.Errorf("Error getting auction round from state")
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
		Unsold: 	quantity,
	}

	// create a composite key for ask using the item and transaction ID
	askKey, err := ctx.GetStub().CreateCompositeKey(askKeyType, []string{auction.ItemSold, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// add to the list of sellers
	sellers := make(map[string]Seller)
	sellers = auction.Sellers
	sellers[askKey] = NewSeller

	newQuantity := 0
	for _, seller := range sellers {
		newQuantity = newQuantity + seller.Quantity
	}
	auction.Quantity = newQuantity

	// If demand <= sold, no need to update the asks
	if auction.Demand <= auction.Sold {
		auction.Sellers = sellers
	} else {

		bidders := make(map[string]Bidder)
		bidders = auction.Bidders

		previousSold := auction.Sold
		newSold := 0
		if auction.Quantity > auction.Demand {
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

	// create a composite key for auction round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	newAuctionJSON, _ := json.Marshal(auction)

	// put update auction in state
	err = ctx.GetStub().PutState(auctionKey, newAuctionJSON)
	if err != nil {
		return fmt.Errorf("failed to update auction: %v", err)
	}

	return nil
}

// DeleteAsk allows the seller of the bid to delete their bid from private data
func (s *SmartContract) DeleteAsk(ctx contractapi.TransactionContextInterface, item string, txID string) error {

	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	collection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// create a composite key using the item and transaction ID
	askKey, err := ctx.GetStub().CreateCompositeKey(askKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// check that the owner is being deleted by the ask owner
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

// NewPublicAsk adds an ask to the public order book. This ensures
// that sellers cannot change their ask during an active auction
func (s *SmartContract) NewPublicAsk(ctx contractapi.TransactionContextInterface, item string, txID string) error {

	// get the implicit collection name using the seller's organization ID
	collection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// create a composite key using the item and transaction ID
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
