/*
SPDX-License-Identifier: Apache-2.0
*/

package auction

import (
	"encoding/json"
	"fmt"
	"strconv"
	"crypto/sha256"
	"bytes"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const privateAskKeyType = "privateAsk"
const publicAskKeyType = "publicAsk"

// Ask is used to sell a certain item. The ask is stored in private data
// of the sellers organization, and identified by the item and transaction id
func (s *SmartContract) Ask(ctx contractapi.TransactionContextInterface, item string) (string, error) {

	// get bid from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("error getting transient: %v", err)
	}

	privateAskJSON, ok := transientMap["privateAsk"]
	if !ok {
		return "", fmt.Errorf("bid key not found in the transient map")
	}

	publicAskJSON, ok := transientMap["publicAsk"]
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
	privateAskKey, err := ctx.GetStub().CreateCompositeKey(privateAskKeyType, []string{item, txID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}

	// put the bid into the organization's implicit data collection
	err = ctx.GetStub().PutPrivateData(collection, privateAskKey, privateAskJSON)
	if err != nil {
		return "", fmt.Errorf("failed to input price into collection: %v", err)
	}

	// create a composite key using the item and transaction ID
	publicAskKey, err := ctx.GetStub().CreateCompositeKey(publicAskKeyType, []string{item, txID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}

	// put the bid into the organization's implicit data collection
	err = ctx.GetStub().PutPrivateData(collection, publicAskKey, publicAskJSON)
	if err != nil {
		return "", fmt.Errorf("failed to input price into collection: %v", err)
	}

	// return the trannsaction ID so that the uset can identify their bid
	return txID, nil
}

// SubmitAsk is used to add an ask to an active auction round
func (s *SmartContract) SubmitAsk(ctx contractapi.TransactionContextInterface, auctionID string, round int, txID string) error {

	// get bid from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	transientAskJSON, ok := transientMap["publicAsk"]
	if !ok {
		return fmt.Errorf("bid key not found in the transient map")
	}

	auction, err := s.QueryAuctionRound(ctx, auctionID, round)
	if err != nil {
		return fmt.Errorf("Error getting auction round from state")
	}

	// create a composite key for bid using the transaction ID
	publicAskKey, err := ctx.GetStub().CreateCompositeKey(publicAskKeyType, []string{auction.ItemSold, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// Check 1: the auction needs to be open for users to add their bid
	status := auction.Status
	if status != "open" {
		return fmt.Errorf("cannot join closed or ended auction")
	}

	// check 3: check that bid has not changed on the public book
	publicAsk, err := s.QueryPublic(ctx, auction.ItemSold, publicAskKeyType, txID)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from public order book: %v", err)
	}

	collection := "_implicit_org_" + publicAsk.Org

	askHash, err := ctx.GetStub().GetPrivateDataHash(collection, publicAskKey)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from collection: %v", err)
	}
	if askHash == nil {
		return fmt.Errorf("bid hash does not exist: %s", askHash)
	}

	hash := sha256.New()
	hash.Write(transientAskJSON)
	calculatedAskJSONHash := hash.Sum(nil)

	// verify that the hash of the passed immutable properties matches the on-chain hash
	if !bytes.Equal(calculatedAskJSONHash, askHash) {
		return fmt.Errorf("hash %x for bid JSON %s does not match hash in auction: %x",
			calculatedAskJSONHash,
			transientAskJSON,
			askHash,
		)
	}

	if !bytes.Equal(publicAsk.Hash, askHash) {
		return fmt.Errorf("Bidder has changed their bid")
	}

	var ask *PublicAsk
	err = json.Unmarshal(transientAskJSON, &ask)
	if err != nil {
		return err
	}

	// store the hash along with the sellers's organization

	newSeller := Seller{
		Seller:   ask.Seller,
		Org:      ask.Org,
		Quantity: ask.Quantity,
		Unsold:   ask.Quantity,
	}

	// add to the list of sellers
	sellers := make(map[string]Seller)
	sellers = auction.Sellers
	sellers[publicAskKey] = newSeller

	newQuantity := 0
	for _, seller := range sellers {
		newQuantity = newQuantity + seller.Quantity
	}
	auction.Quantity = newQuantity
	auction.Sellers = sellers

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
	privateAskKey, err := ctx.GetStub().CreateCompositeKey(privateAskKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// check that the owner is being deleted by the ask owner
	err = s.checkAskOwner(ctx, collection, privateAskKey)
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelPrivateData(collection, privateAskKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", privateAskKey, err)
	}

	// create a composite key using the item and transaction ID
	publicAskKey, err := ctx.GetStub().CreateCompositeKey(publicAskKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().DelPrivateData(collection, publicAskKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", publicAskKey, err)
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
	askKey, err := ctx.GetStub().CreateCompositeKey(publicAskKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	hash, err := ctx.GetStub().GetPrivateDataHash(collection, askKey)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from collection: %v", err)
	}
	if hash == nil {
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
		Hash: hash,
	}

	publicAskJSON, _ := json.Marshal(publicAsk)

	// put the ask hash of the bid in the public order book
	err = ctx.GetStub().PutState(askKey, publicAskJSON)
	if err != nil {
		return fmt.Errorf("failed to input bid into state: %v", err)
	}

	return nil
}
