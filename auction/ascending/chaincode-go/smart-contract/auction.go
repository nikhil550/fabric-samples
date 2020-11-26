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

type SmartContract struct {
	contractapi.Contract
}

// Auction data
type Auction struct {
	Type     string            `json:"objectType"`
	Round    int               `json:"round"`
	Status   string            `json:"status"`
	ItemSold string            `json:"item"`
	Orgs     []string          `json:"organizations"`
	Price    int               `json:"price"`
	Quantity int               `json:"quantity"`
	Demand   int               `json:"demand"`
	Sellers  map[string]Seller `json:"sellers"`
	Bidders  map[string]Bidder `json:"bidders"`
	Winners  map[string]Winner `json:"winners"`
}

// Bid is the structure of a bid in private state
type Bid struct {
	Type     string `json:"objectType"`
	Quantity int    `json:"quantity"`
	Org      string `json:"org"`
	Buyer    string `json:"buyer"`
	Price    int    `json:"price"`
}

// Ask is the structure of a bid in private state
type Ask struct {
	Type     string `json:"objectType"`
	Quantity int    `json:"quantity"`
	Org      string `json:"org"`
	Seller   string `json:"seller"`
	Price    int    `json:"price"`
}

// Bidder is the structure that lives on the auction
type Bidder struct {
	Buyer    string `json:"buyer"`
	Org      string `json:"org"`
	Quantity int    `json:"quantity"`
}

// Seller is the structure that lives on the auction
type Seller struct {
	Seller   string `json:"seller"`
	Org      string `json:"org"`
	Quantity int    `json:"quantity"`
}

// Winners stores the winners of the auction round
type Winner struct {
	Buyer       string `json:"buyer"`
	QuantityBid int    `json:"quantityBid"`
	QuantityWon int    `json:"quantityWon"`
}

// CreateAuction creates on auction on the public channel. Each auction round is
// stored as a seperate key in the world state
func (s *SmartContract) CreateAuction(ctx contractapi.TransactionContextInterface, auctionID string, itemsold string) error {

	// get org of submitting client
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	// Create auction
	sellers := make(map[string]Seller)
	bidders := make(map[string]Bidder)
	winners := make(map[string]Winner)

	auction := Auction{
		Type:     "auction",
		Round:    0,
		Status:   "open",
		ItemSold: itemsold,
		Orgs:     []string{clientOrgID},
		Quantity: 0,
		Demand:   0,
		Price:    0,
		Sellers:  sellers,
		Bidders:  bidders,
		Winners:  winners,
	}

	auctionBytes, err := json.Marshal(auction)
	if err != nil {
		return err
	}

	// create a composite key using the round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(auction.Round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// put auction into state
	err = ctx.GetStub().PutState(auctionKey, auctionBytes)
	if err != nil {
		return fmt.Errorf("failed to put auction in public data: %v", err)
	}

	// set the seller of the auction as an endorser
	err = setAssetStateBasedEndorsement(ctx, auctionID, clientOrgID)
	if err != nil {
		return fmt.Errorf("failed setting state based endorsement for new organization: %v", err)
	}

	return nil
}

func (s *SmartContract) CreateNewRound(ctx contractapi.TransactionContextInterface, auctionID string, price int) error {

	auctionBytes, err := ctx.GetStub().GetState(auctionID)
	if err != nil {
		return fmt.Errorf("failed to get auction %v: %v", auctionID, err)
	}

	if auctionBytes == nil {
		return fmt.Errorf("Auction interest object %v not found", auctionID)
	}

	var auctionJSON Auction
	err = json.Unmarshal(auctionBytes, &auctionJSON)
	if err != nil {
		return fmt.Errorf("failed to create auction object JSON: %v", err)
	}

	// TODO: logic on when to create a new round
	// check 1: Demand > supply
	// check 2: A there is a higher ask in our orgs book

	auctionJSON.Round = auctionJSON.Round + 1

	auctionJSON.Price = price

	newRound, _ := json.Marshal(auctionJSON)

	// create a composite key using the transaction ID
	newRoundKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(auctionJSON.Round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().PutState(newRoundKey, newRound)
	if err != nil {
		return fmt.Errorf("failed to close auction: %v", err)
	}

	return nil
}

// SubmitBid is used by the bidder to add the hash of that bid stored in private data to the
// auction. Note that this function alters the auction in private state, and needs
// to meet the auction endorsement policy. Transaction ID is used identify the bid
func (s *SmartContract) SubmitBid(ctx contractapi.TransactionContextInterface, auctionID string, round int, quantity int, txID string) error {

	// get identity of submitting client
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	// get the org of the subitting bidder
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	// create a composite for auction using the round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// get the auction from state
	auctionBytes, err := ctx.GetStub().GetState(auctionKey)

	var auctionJSON Auction

	if auctionBytes == nil {
		return fmt.Errorf("Auction not found: %v", auctionID)
	}
	err = json.Unmarshal(auctionBytes, &auctionJSON)
	if err != nil {
		return fmt.Errorf("failed to create auction object JSON: %v", err)
	}

	// the auction needs to be open for users to add their bid
	Status := auctionJSON.Status
	if Status != "open" {
		return fmt.Errorf("cannot join closed or ended auction")
	}

	NewBidder := Bidder{
		Buyer:    clientID,
		Org:      clientOrgID,
		Quantity: quantity,
	}

	// create a composite key for bid using the transaction ID
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{auctionJSON.ItemSold, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	bidders := make(map[string]Bidder)
	bidders = auctionJSON.Bidders
	bidders[bidKey] = NewBidder
	auctionJSON.Bidders = bidders

	auctionJSON.Demand = auctionJSON.Demand + NewBidder.Quantity

	// Add the bidding organization to the list of participating organization's if it is not already
	Orgs := auctionJSON.Orgs
	if !(contains(Orgs, clientOrgID)) {
		newOrgs := append(Orgs, clientOrgID)
		auctionJSON.Orgs = newOrgs

		err = addAssetStateBasedEndorsement(ctx, auctionKey, clientOrgID)
		if err != nil {
			return fmt.Errorf("failed setting state based endorsement for new organization: %v", err)
		}
	}

	newAuctionBytes, _ := json.Marshal(auctionJSON)

	err = ctx.GetStub().PutState(auctionKey, newAuctionBytes)
	if err != nil {
		return fmt.Errorf("failed to update auction: %v", err)
	}

	return nil
}

func (s *SmartContract) SubmitAsk(ctx contractapi.TransactionContextInterface, auctionID string, round int, quantity int, txID string) error {

	// get identity of submitting client
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	// get the org of the subitting bidder
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	// create a composite for auction using the round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// get the auction from state
	auctionBytes, err := ctx.GetStub().GetState(auctionKey)

	var auctionJSON Auction

	if auctionBytes == nil {
		return fmt.Errorf("Auction not found: %v", auctionID)
	}
	err = json.Unmarshal(auctionBytes, &auctionJSON)
	if err != nil {
		return fmt.Errorf("failed to create auction object JSON: %v", err)
	}

	// the auction needs to be open for users to add their bid
	Status := auctionJSON.Status
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
	askKey, err := ctx.GetStub().CreateCompositeKey(askKeyType, []string{auctionJSON.ItemSold, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	sellers := make(map[string]Seller)
	sellers = auctionJSON.Sellers
	sellers[askKey] = NewSeller
	auctionJSON.Sellers = sellers

	auctionJSON.Quantity = auctionJSON.Quantity + NewSeller.Quantity

	// Add the bidding organization to the list of participating organization's if it is not already
	Orgs := auctionJSON.Orgs
	if !(contains(Orgs, clientOrgID)) {
		newOrgs := append(Orgs, clientOrgID)
		auctionJSON.Orgs = newOrgs

		err = addAssetStateBasedEndorsement(ctx, auctionKey, clientOrgID)
		if err != nil {
			return fmt.Errorf("failed setting state based endorsement for new organization: %v", err)
		}
	}

	newAuctionBytes, _ := json.Marshal(auctionJSON)

	err = ctx.GetStub().PutState(auctionKey, newAuctionBytes)
	if err != nil {
		return fmt.Errorf("failed to update auction: %v", err)
	}

	return nil
}

// CloseAuction can be used by the seller to close the auction. This prevents
// bids from being added to the auction, and allows users to reveal their bid
func (s *SmartContract) CloseAuction(ctx contractapi.TransactionContextInterface, auctionID string, round int) error {

	// create a composite key using the round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	auctionBytes, err := ctx.GetStub().GetState(auctionKey)
	if err != nil {
		return fmt.Errorf("failed to get auction %v: %v", auctionKey, err)
	}

	if auctionBytes == nil {
		return fmt.Errorf("Auction interest object %v not found", auctionKey)
	}

	var auctionJSON Auction
	err = json.Unmarshal(auctionBytes, &auctionJSON)
	if err != nil {
		return fmt.Errorf("failed to create auction object JSON: %v", err)
	}

	auctionJSON.Status = string("closed")

	closedAuction, _ := json.Marshal(auctionJSON)

	err = ctx.GetStub().PutState(auctionKey, closedAuction)
	if err != nil {
		return fmt.Errorf("failed to close auction: %v", err)
	}

	return nil
}

// EndAuction both changes the auction status to closed and calculates the winners
// of the auction
func (s *SmartContract) EndAuction(ctx contractapi.TransactionContextInterface, auctionID string, round int) error {

	// create a composite key using the round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	auctionBytes, err := ctx.GetStub().GetState(auctionKey)
	if err != nil {
		return fmt.Errorf("failed to get auction %v: %v", auctionKey, err)
	}

	if auctionBytes == nil {
		return fmt.Errorf("Auction interest object %v not found", auctionKey)
	}

	var auctionJSON Auction
	err = json.Unmarshal(auctionBytes, &auctionJSON)
	if err != nil {
		return fmt.Errorf("failed to create auction object JSON: %v", err)
	}

	Status := auctionJSON.Status
	if Status != "closed" {
		return fmt.Errorf("Can only end a closed auction")
	}

	// check if there is a winning bid that has yet to be revealed
	//	err = queryAllBids(ctx, auctionJSON.Price, auctionJSON.Bidders)
	//	if err != nil {
	//		return fmt.Errorf("Cannot close auction: %v", err)
	//	}

	auctionJSON.Status = string("ended")

	closedAuction, _ := json.Marshal(auctionJSON)

	err = ctx.GetStub().PutState(auctionID, closedAuction)
	if err != nil {
		return fmt.Errorf("failed to close auction: %v", err)
	}
	return nil
}
