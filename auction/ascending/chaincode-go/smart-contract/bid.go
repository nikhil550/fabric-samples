/*
SPDX-License-Identifier: Apache-2.0
*/

package auction

import (
	//"bytes"
	"encoding/json"
	"fmt"
	"crypto/sha256"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const bidKeyType = "bid"
const priceKeyType = "price"
const askKeyType = "ask"
const item = "widgets"

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

	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed getting the client's MSPID: %v", err)
	}

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

	hash := sha256.New()
	hash.Write(BidJSON)
	BidJSONhash := hash.Sum(nil)

	// store the hash along with the bidder's organization in the public order book
	publicBidJSON := BidAskHash{
		Org:  clientOrgID,
		Hash: fmt.Sprintf("%x", BidJSONhash),
	}

	publicBidBytes, _ := json.Marshal(publicBidJSON)

	// put the bid hash of the bid in the public order book
	err = ctx.GetStub().PutState(bidKey, publicBidBytes)
	if err != nil {
		return "", fmt.Errorf("failed to input bid into state: %v", err)
	}

	// set the seller of the auction as an endorser
	err = setAssetStateBasedEndorsement(ctx, bidKey, clientOrgID)
	if err != nil {
		return "", fmt.Errorf("failed setting state based endorsement for new organization: %v", err)
	}

	// return the trannsaction ID so that the uset can identify their bid
	return txID, nil
}

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

	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed getting the client's MSPID: %v", err)
	}

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

	hash := sha256.New()
	hash.Write(askJSON)
	AskJSONhash := hash.Sum(nil)

	// store the hash along with the seller's organization in the public order book
	publicAskJSON := BidAskHash{
		Org:  clientOrgID,
		Hash: fmt.Sprintf("%x", AskJSONhash),
	}

	publicAskBytes, _ := json.Marshal(publicAskJSON)

	// put the ask hash of the bid in the public order book
	err = ctx.GetStub().PutState(askKey, publicAskBytes)
	if err != nil {
		return "", fmt.Errorf("failed to input bid into state: %v", err)
	}

	// set the seller of the auction as an endorser
	err = setAssetStateBasedEndorsement(ctx, askKey, clientOrgID)
	if err != nil {
		return "", fmt.Errorf("failed setting state based endorsement for new organization: %v", err)
	}

	// return the trannsaction ID so that the uset can identify their bid
	return txID, nil
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

	err = checkBidOwner(ctx, collection, bidKey)
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelPrivateData(collection, bidKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", bidKey, err)
	}

	err = ctx.GetStub().DelState(bidKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", bidKey, err)
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

	err = checkAskOwner(ctx, collection, askKey)
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelPrivateData(collection, askKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", askKey, err)
	}

	err = ctx.GetStub().DelState(askKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", askKey, err)
	}

	return nil
}


func checkBidOwner(ctx contractapi.TransactionContextInterface, collection string, bidKey string) error {

	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	bidJSON, err := ctx.GetStub().GetPrivateData(collection, bidKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", bidKey, err)
	}
	if bidJSON == nil {
		return fmt.Errorf("bid %v does not exist", bidKey)
	}

	var bid *Bid
	err = json.Unmarshal(bidJSON, &bid)
	if err != nil {
		return err
	}

	// check that the client querying the bid is the bid owner
	if bid.Buyer != clientID {
		return fmt.Errorf("Permission denied, client id %v is not the owner of the bid", clientID)
	}

	return nil
}

func checkAskOwner(ctx contractapi.TransactionContextInterface, collection string, askKey string) error {

	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	askJSON, err := ctx.GetStub().GetPrivateData(collection, askKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", askKey, err)
	}
	if askJSON == nil {
		return fmt.Errorf("bid %v does not exist", askKey)
	}

	var ask *Ask
	err = json.Unmarshal(askJSON, &ask)
	if err != nil {
		return err
	}

	// check that the client querying the bid is the bid owner
	if ask.Seller != clientID {
		return fmt.Errorf("Permission denied, client id %v is not the owner of the ask", clientID)
	}

	return nil
}
