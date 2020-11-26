/*
SPDX-License-Identifier: Apache-2.0
*/

package auction

import (
	"encoding/json"
	"fmt"
	"strconv"

	//	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// QueryAuction allows all members of the channel to read a public auction
func (s *SmartContract) QueryAuction(ctx contractapi.TransactionContextInterface, auctionID string) ([]*Auction, error) {

	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("auction", []string{auctionID})
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var auctions []*Auction
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var auction Auction
		err = json.Unmarshal(queryResponse.Value, &auction)
		if err != nil {
			return nil, err
		}

		auctions = append(auctions, &auction)
	}

	return auctions, nil
}

// QueryAuction allows all members of the channel to read a public auction
func (s *SmartContract) QueryAuctionRound(ctx contractapi.TransactionContextInterface, auctionID string, round int) (*Auction, error) {

	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	auctionJSON, err := ctx.GetStub().GetState(auctionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get auction object %v: %v", auctionID, err)
	}
	if auctionJSON == nil {
		return nil, fmt.Errorf("auction does not exist")
	}

	var auction *Auction
	err = json.Unmarshal(auctionJSON, &auction)
	if err != nil {
		return nil, err
	}

	return auction, nil
}

// QueryBid allows the submitter of the bid to read their bid from public state
func (s *SmartContract) QueryBid(ctx contractapi.TransactionContextInterface, item string, txID string) (*Bid, error) {

	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return nil, fmt.Errorf("failed to get client identity %v", err)
	}

	collection, err := getCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{item, txID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	bidJSON, err := ctx.GetStub().GetPrivateData(collection, bidKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid %v: %v", bidKey, err)
	}
	if bidJSON == nil {
		return nil, fmt.Errorf("bid %v does not exist", bidKey)
	}

	var bid *Bid
	err = json.Unmarshal(bidJSON, &bid)
	if err != nil {
		return nil, err
	}

	// check that the client querying the bid is the bid owner
	if bid.Buyer != clientID {
		return nil, fmt.Errorf("Permission denied, client id %v is not the owner of the bid", clientID)
	}

	return bid, nil
}

// QueryBid allows the submitter of the bid to read their bid from public state
func (s *SmartContract) QueryAsk(ctx contractapi.TransactionContextInterface, item string, txID string) (*Ask, error) {

	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return nil, fmt.Errorf("failed to get client identity %v", err)
	}

	collection, err := getCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	askKey, err := ctx.GetStub().CreateCompositeKey(askKeyType, []string{item, txID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	askJSON, err := ctx.GetStub().GetPrivateData(collection, askKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid %v: %v", askKey, err)
	}
	if askJSON == nil {
		return nil, fmt.Errorf("bid %v does not exist", askKey)
	}

	var ask *Ask
	err = json.Unmarshal(askJSON, &ask)
	if err != nil {
		return nil, err
	}

	// check that the client querying the bid is the bid owner
	if ask.Seller != clientID {
		return nil, fmt.Errorf("Permission denied, client id %v is not the seller", clientID)
	}

	return ask, nil
}

// GetID is an internal helper function to allow users to get their identity
func (s *SmartContract) GetID(ctx contractapi.TransactionContextInterface) (string, error) {

	// Get the MSP ID of submitting client identity
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed to get verified MSPID: %v", err)
	}

	return clientID, nil
}
