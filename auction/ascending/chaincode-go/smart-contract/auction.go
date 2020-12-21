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

// BidAskHash is the structure of a private bid or ask in the public order book
type BidAskHash struct {
	Org  string `json:"org"`
	Hash string `json:"hash"`
}

// Bidder is the structure that lives on the auction
type Bidder struct {
	Buyer    string `json:"buyer"`
	Org      string `json:"org"`
	Quantity int    `json:"quantityBid"`
	Won      int    `json:"quantityWon"`
}

// Seller is the structure that lives on the auction
type Seller struct {
	Seller   string `json:"seller"`
	Org      string `json:"org"`
	Quantity int    `json:"quantity"`
}

// CreateAuction creates on auction on the public channel. Each auction round is
// stored as a seperate key in the world state
func (s *SmartContract) CreateAuction(ctx contractapi.TransactionContextInterface, auctionID string, itemSold string, reservePrice int) error {

	// get org of submitting client
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	sellers := make(map[string]Seller)
	bidders := make(map[string]Bidder)

	// Check 1: check if there is an ask from your org that is lower than the reserve price
	err = queryAllAsks(ctx, reservePrice, itemSold, sellers)
	if err != nil {
		return fmt.Errorf("Cannot close auction: %v", err)
	}

	// Create auction

	auction := Auction{
		Type:     "auction",
		Round:    0,
		Status:   "open",
		ItemSold: itemSold,
		Orgs:     []string{clientOrgID},
		Quantity: 0,
		Demand:   0,
		Price:    reservePrice,
		Sellers:  sellers,
		Bidders:  bidders,
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

func (s *SmartContract) CreateNewRound(ctx contractapi.TransactionContextInterface, auctionID string, newRound int, price int) error {

	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the client's MSPID: %v", err)
	}

	// check 1: was there a previous round

	previousRound := newRound - 1

	auction, err := s.QueryAuctionRound(ctx, auctionID, previousRound)
	if err != nil {
		return fmt.Errorf("Cannot create round until previous round is created")
	}

	// check 2: confirm that Demand > Supply for the previous round before creating a new round

	nextRound := auction.Round + 1

	if auction.Quantity > auction.Demand {
		return fmt.Errorf("Cannot create new round: demand is not yet greater than supply")
	}

	// check 3: check if there is a lower ask in our orgs book

	err = queryAllAsks(ctx, auction.Price, auction.ItemSold, auction.Sellers)
	if err != nil {
		return fmt.Errorf("Cannot close auction: %v", err)
	}

	// Becuase both checks passed, create a new round

	auction.Round = nextRound

	auction.Price = price

	newAuctionRound, err := json.Marshal(auction)
	if err != nil {
		return err
	}
	// create a composite key using the transaction ID
	newAuctionRoundKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(auction.Round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().PutState(newAuctionRoundKey, newAuctionRound)
	if err != nil {
		return fmt.Errorf("failed to close auction: %v", err)
	}

	// set the seller of the auction as an endorser
	err = setAssetStateBasedEndorsement(ctx, auctionID, clientOrgID)
	if err != nil {
		return fmt.Errorf("failed setting state based endorsement for new organization: %v", err)
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

	if previousRound != 0 {

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
	// create new bid

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
		for _, bidder := range bidders {
			bidder.Won = bidder.Quantity
		}
	} else if auction.Quantity == 0 {
		for _, bidder := range bidders {
			bidder.Won = 0
		}
	}	else {
		for _, bidder := range bidders {
			bidder.Won = (bidder.Quantity * auction.Demand) / auction.Quantity
		}
	}
	auction.Bidders = bidders

	// create a composite for auction using the round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// Add the bidding organization to the list of participating organization's if it is not already
	Orgs := auction.Orgs
	if !(contains(Orgs, clientOrgID)) {
		newOrgs := append(Orgs, clientOrgID)
		auction.Orgs = newOrgs

		err = addAssetStateBasedEndorsement(ctx, auctionKey, clientOrgID)
		if err != nil {
			return fmt.Errorf("failed setting state based endorsement for new organization: %v", err)
		}
	}

	newAuctionBytes, _ := json.Marshal(auction)

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
		for _, bidder := range bidders {
			bidder.Won = bidder.Quantity
		}
	}	else {
		for _, bidder := range bidders {
			bidder.Won = (bidder.Quantity * auction.Demand) / auction.Quantity
		}
	}

	auction.Bidders = bidders

	// create a composite for auction using the round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// Add the selling organization to the list of participating organization's if it is not already
	Orgs := auction.Orgs
	if !(contains(Orgs, clientOrgID)) {
		newOrgs := append(Orgs, clientOrgID)
		auction.Orgs = newOrgs

		err = addAssetStateBasedEndorsement(ctx, auctionKey, clientOrgID)
		if err != nil {
			return fmt.Errorf("failed setting state based endorsement for new organization: %v", err)
		}
	}

	newAuctionBytes, _ := json.Marshal(auction)

	err = ctx.GetStub().PutState(auctionKey, newAuctionBytes)
	if err != nil {
		return fmt.Errorf("failed to update auction: %v", err)
	}

	return nil
}

// CloseAuction can be used by the seller to close the auction. This prevents
// bids from being added to the auction, and allows users to reveal their bid
func (s *SmartContract) CloseAuction(ctx contractapi.TransactionContextInterface, auctionID string, round int) error {

	auction, err := s.QueryAuctionRound(ctx, auctionID, round)
	if err != nil {
		return err
	}

	Status := auction.Status
	if Status != "open" {
		return fmt.Errorf("Can only close an open auction")
	}

	// check if there is a winning bid that has yet to be revealed
	err = queryAllBids(ctx, auction.Price,  auction.ItemSold, auction.Bidders)
	if err != nil {
		return fmt.Errorf("Cannot close auction: %v", err)
	}

	auction.Status = string("closed")

	closedAuction, _ := json.Marshal(auction)

	// create a composite key using the round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().PutState(auctionKey, closedAuction)
	if err != nil {
		return fmt.Errorf("failed to close auction: %v", err)
	}

	return nil
}

// EndAuction both changes the auction status to closed and calculates the winners
// of the auction
func (s *SmartContract) EndAuction(ctx contractapi.TransactionContextInterface, auctionID string, round int) error {

	auction, err := s.QueryAuctionRound(ctx, auctionID, round)
	if err != nil {
		return err
	}

	Status := auction.Status
	if Status != "closed" {
		return fmt.Errorf("Can only end a closed auction")
	}

	auction.Status = string("ended")

	closedAuction, _ := json.Marshal(auction)

	// create a composite key using the round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().PutState(auctionKey, closedAuction)
	if err != nil {
		return fmt.Errorf("failed to close auction: %v", err)
	}
	return nil
}
