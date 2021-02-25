/*
SPDX-License-Identifier: Apache-2.0
*/

package auction

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// BidReturn is the data type returned to an auction admin
type BidReturn struct {
	ID  string `json:"id"`
	Bid *PrivateBid   `json:"bid"`
}

// AskReturn is the data type returned to an auction admin
type AskReturn struct {
	ID  string `json:"id"`
	Ask *PrivateAsk   `json:"ask"`
}

// QueryAuction allows all members of the channel to read all rounds of a public auction
func (s *SmartContract) QueryAuction(ctx contractapi.TransactionContextInterface, auctionID string) ([]*AuctionRound, error) {

	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("auction", []string{auctionID})
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var auctionRounds []*AuctionRound
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var auctionRound AuctionRound
		err = json.Unmarshal(queryResponse.Value, &auctionRound)
		if err != nil {
			return nil, err
		}

		auctionRounds = append(auctionRounds, &auctionRound)
	}

	return auctionRounds, nil
}

// QueryAuctionRound allows all members of the channel to read a public auction round
func (s *SmartContract) QueryAuctionRound(ctx contractapi.TransactionContextInterface, auctionID string, round int) (*AuctionRound, error) {

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

	var auctionRound *AuctionRound
	err = json.Unmarshal(auctionJSON, &auctionRound)
	if err != nil {
		return nil, err
	}

	return auctionRound, nil
}

// QueryBid allows the submitter of the bid or an auction admin to read their bid from private state
func (s *SmartContract) QueryBid(ctx contractapi.TransactionContextInterface, item string, txID string) (*PrivateBid, error) {

	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	collection, err := getCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	bidKey, err := ctx.GetStub().CreateCompositeKey(privateBidKeyType, []string{item, txID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	// only the bid owner or the auction admin can read a bid
	err = s.checkBidOwner(ctx, collection, bidKey)
	if err != nil {
		err = ctx.GetClientIdentity().AssertAttributeValue("role", "auctionAdmin")
		if err != nil {
			return nil, fmt.Errorf("submitting client needs to be the bid owner or an auction admin")
		}
	}

	bidJSON, err := ctx.GetStub().GetPrivateData(collection, bidKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid %v: %v", bidKey, err)
	}
	if bidJSON == nil {
		return nil, fmt.Errorf("bid %v does not exist", bidKey)
	}

	var bid *PrivateBid
	err = json.Unmarshal(bidJSON, &bid)
	if err != nil {
		return nil, err
	}

	return bid, nil
}

// QueryAsk allows a seller or an auction admin to read their bid from private state
func (s *SmartContract) QueryAsk(ctx contractapi.TransactionContextInterface, item string, txID string) (*PrivateAsk, error) {

	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	collection, err := getCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	askKey, err := ctx.GetStub().CreateCompositeKey(privateAskKeyType, []string{item, txID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	err = s.checkAskOwner(ctx, collection, askKey)
	if err != nil {
		err = ctx.GetClientIdentity().AssertAttributeValue("role", "auctionAdmin")
		if err != nil {
			return nil, fmt.Errorf("submitting client needs to be the ask owner or an auction admin")
		}
	}

	askJSON, err := ctx.GetStub().GetPrivateData(collection, askKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid %v: %v", askKey, err)
	}
	if askJSON == nil {
		return nil, fmt.Errorf("ask %v does not exist", askKey)
	}

	var ask *PrivateAsk
	err = json.Unmarshal(askJSON, &ask)
	if err != nil {
		return nil, err
	}

	return ask, nil
}

// QueryBids returns all bids from a private data collection for a certain item.
// this function is used by auction admins to add bids to a open auction
func (s *SmartContract) QueryBids(ctx contractapi.TransactionContextInterface, item string) ([]BidReturn, error) {

	// the function can only be used by an auction admin
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "auctionAdmin")
	if err != nil {
		return nil, fmt.Errorf("submitting client needs to be an auction admin")
	}

	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	collection, err := getCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// return bids using the item
	resultsIterator, err := ctx.GetStub().GetPrivateDataByPartialCompositeKey(collection, privateBidKeyType, []string{item})
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// return the bid and the transaction id, so that the bid can be submitted
	var bidReturns []BidReturn
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		_, keyParts, Err := ctx.GetStub().SplitCompositeKey(queryResponse.Key)
		if Err != nil {
			return nil, err
		}

		txID := keyParts[1]

		var bid *PrivateBid
		err = json.Unmarshal(queryResponse.Value, &bid)
		if err != nil {
			return nil, err
		}

		bidReturn := BidReturn{
			ID:  txID,
			Bid: bid,
		}

		bidReturns = append(bidReturns, bidReturn)
	}

	return bidReturns, nil
}

// QueryAsks returns all asks from a private data collection for a certain item.
// this function is used by auction admins to add asks to a open auction
func (s *SmartContract) QueryAsks(ctx contractapi.TransactionContextInterface, item string) ([]AskReturn, error) {

	err := ctx.GetClientIdentity().AssertAttributeValue("role", "auctionAdmin")
	if err != nil {
		return nil, fmt.Errorf("submitting client needs to be an auction admin")
	}

	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	collection, err := getCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// return ask using the item
	resultsIterator, err := ctx.GetStub().GetPrivateDataByPartialCompositeKey(collection, privateAskKeyType, []string{item})
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// return the ask and the transaction id, so that the bid can be submitted
	var askReturns []AskReturn
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		_, keyParts, Err := ctx.GetStub().SplitCompositeKey(queryResponse.Key)
		if Err != nil {
			return nil, Err
		}

		txID := keyParts[1]

		var ask *PrivateAsk
		err = json.Unmarshal(queryResponse.Value, &ask)
		if err != nil {
			return nil, err
		}

		askReturn := AskReturn{
			ID:  txID,
			Ask: ask,
		}

		askReturns = append(askReturns, askReturn)

	}

	return askReturns, nil
}

// checkForHigherBid is an internal function that is used to determine if
// there is a higher bid that has yet to be added to an auction round
func checkForHigherBid(ctx contractapi.TransactionContextInterface, auctionPrice int, item string, bidders map[string]Bidder) error {

	// Get MSP ID of peer org
	peerMSPID, err := shim.GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the peer's MSPID: %v", err)
	}

	var error error
	error = nil

	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(privateBidKeyType, []string{item})
	if err != nil {
		return err
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return err
		}

		var publicBid BidAskHash
		err = json.Unmarshal(queryResponse.Value, &publicBid)
		if err != nil {
			return err
		}

		bidKey := queryResponse.Key

		if _, bidInAuction := bidders[bidKey]; bidInAuction {

			//bid is already in the auction, no action to take

		} else {

			collection := "_implicit_org_" + publicBid.Org

			if publicBid.Org == peerMSPID {

				bidJSON, err := ctx.GetStub().GetPrivateData(collection, bidKey)
				if err != nil {
					return fmt.Errorf("failed to get bid %v: %v", bidKey, err)
				}
				if bidJSON == nil {
					return fmt.Errorf("bid %v does not exist", bidKey)
				}

				var bid *PrivateBid
				err = json.Unmarshal(bidJSON, &bid)
				if err != nil {
					return err
				}

				if bid.Price >= auctionPrice {
					error = fmt.Errorf("Cannot close auction round, bidder has a higher price: %v", err)
				}

			} else {

				hash, err := ctx.GetStub().GetPrivateDataHash(collection, bidKey)
				if err != nil {
					return fmt.Errorf("failed to read bid hash from collection: %v", err)
				}
				if hash == nil {
					return fmt.Errorf("bid hash does not exist: %s", bidKey)
				}
			}
		}
	}

	return error
}

// checkForLowerAsk is an internal function that is used to determine
// is there is a lower ask that has not yet been added to the round
func checkForLowerAsk(ctx contractapi.TransactionContextInterface, auctionPrice int, item string, sellers map[string]Seller) error {

	// Get MSP ID of peer org
	peerMSPID, err := shim.GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the peer's MSPID: %v", err)
	}

	var error error
	error = nil

	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(privateAskKeyType, []string{item})
	if err != nil {
		return err
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return err
		}

		var publicAsk BidAskHash
		err = json.Unmarshal(queryResponse.Value, &publicAsk)
		if err != nil {
			return err
		}

		askKey := queryResponse.Key

		if _, askInAuction := sellers[askKey]; askInAuction {

			//ask is already in the auction, no action to take

		} else {

			collection := "_implicit_org_" + publicAsk.Org

			if publicAsk.Org == peerMSPID {

				askJSON, err := ctx.GetStub().GetPrivateData(collection, askKey)
				if err != nil {
					return fmt.Errorf("failed to get bid %v: %v", askKey, err)
				}
				if askJSON == nil {
					return fmt.Errorf("ask %v does not exist", askKey)
				}

				var ask *PrivateAsk
				err = json.Unmarshal(askJSON, &ask)
				if err != nil {
					return err
				}

				if ask.Price <= auctionPrice {
					error = fmt.Errorf("Cannot close auction round, seller has a lower price: %v", err)
				}

			} else {

				hash, err := ctx.GetStub().GetPrivateDataHash(collection, askKey)
				if err != nil {
					return fmt.Errorf("failed to read bid hash from collection: %v", err)
				}
				if hash == nil {
					return fmt.Errorf("bid hash does not exist: %s", askKey)
				}
			}
		}
	}

	return error
}

// QueryPublic allows you to read the public hash on the order book
func (s *SmartContract) QueryPublic(ctx contractapi.TransactionContextInterface, item string, askSell string, txID string) (*BidAskHash, error) {

	bidAskKey, err := ctx.GetStub().CreateCompositeKey(askSell, []string{item, txID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	bidAskJSON, err := ctx.GetStub().GetState(bidAskKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid %v: %v", bidAskKey, err)
	}
	if bidAskJSON == nil {
		return nil, fmt.Errorf("bid or ask %v does not exist", bidAskKey)
	}

	var hash *BidAskHash
	err = json.Unmarshal(bidAskJSON, &hash)
	if err != nil {
		return nil, err
	}

	return hash, nil
}
