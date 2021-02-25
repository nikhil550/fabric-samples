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

// Auction Round is the structure of a bid in public state
type AuctionRound struct {
	Type     string            `json:"objectType"`
	ID       string            `json:"id"`
	Round    int               `json:"round"`
	Status   string            `json:"status"`
	ItemSold string            `json:"item"`
	Price    int               `json:"price"`
	Quantity int               `json:"quantity"`
	Sold     int               `json:"sold"`
	Demand   int               `json:"demand"`
	Sellers  map[string]Seller `json:"sellers"`
	Bidders  map[string]Bidder `json:"bidders"`
}

// PrivateBid is the structure of a bid in private state
type PrivateBid struct {
	Type     string `json:"objectType"`
	Quantity int    `json:"quantity"`
	Org      string `json:"org"`
	Buyer    string `json:"buyer"`
	Price    int    `json:"price"`
}

// Bid is the structure of a bid that will be made public
type PublicBid struct {
	Type     string `json:"objectType"`
	Quantity int    `json:"quantity"`
	Org      string `json:"org"`
	Buyer    string `json:"buyer"`
	Price    int    `json:"price"`
}

// PrivateAsk is the structure of a bid in private state
type PrivateAsk struct {
	Type     string `json:"objectType"`
	Quantity int    `json:"quantity"`
	Org      string `json:"org"`
	Seller   string `json:"seller"`
	Price    int    `json:"price"`
}

// PublicAsk is the structure of a bid in public state
type PublicAsk struct {
	Type     string `json:"objectType"`
	Quantity int    `json:"quantity"`
	Org      string `json:"org"`
	Seller   string `json:"seller"`
}


// BidAskHash is the structure of a private bid or ask in the public order book
type BidAskHash struct {
	Org  string `json:"org"`
	Hash []byte `json:"hash"`
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
	Sold     int    `json:"sold"`
	Unsold   int    `json:"unsold"`
}

// incrementAmount is the price increase of each new round of the auction
const incrementAmount = 5

// CreateAuction creates a new auction on the public channel.
// Each round of teh auction is stored as a seperate key in the world state
func (s *SmartContract) CreateAuction(ctx contractapi.TransactionContextInterface, auctionID string, itemSold string, reservePrice int) error {

	existingAuction, err := s.QueryAuction(ctx, auctionID)
	if existingAuction != nil {
		return fmt.Errorf("Cannot create new auction: auction already exists")
	}

	sellers := make(map[string]Seller)
	bidders := make(map[string]Bidder)

	// Check if there is an ask from your org that is lower
	// than the reserve price of the first round before creating the auction
	err = checkForLowerAsk(ctx, reservePrice, itemSold, sellers)
	if err != nil {
		return fmt.Errorf("seller has lower ask, cannot open a new auction at this price: %v", err)
	}

	// Create the first round of the auction
	auction := AuctionRound{
		Type:     "auction",
		ID:       auctionID,
		Round:    0,
		Status:   "open",
		ItemSold: itemSold,
		Quantity: 0,
		Demand:   0,
		Price:    reservePrice,
		Sellers:  sellers,
		Bidders:  bidders,
	}

	auctionJSON, err := json.Marshal(auction)
	if err != nil {
		return err
	}

	// create a composite key for the auction round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(auction.Round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// put auction round into state
	err = ctx.GetStub().PutState(auctionKey, auctionJSON)
	if err != nil {
		return fmt.Errorf("failed to put auction in public data: %v", err)
	}

	// create an event to notify buyers and sellers of a new auction
	err = ctx.GetStub().SetEvent("CreateAuction", []byte(auctionID))
	if err != nil {
		return fmt.Errorf("event failed to register: %v", err)
	}

	return nil
}

// CreateNewRound creates a new round of the auction. The new round has a seperate key
// in world state. Bidders and sellers have the abiltiy to join the round at the
// new price
func (s *SmartContract) CreateNewRound(ctx contractapi.TransactionContextInterface, auctionID string, newRound int) error {

	// checks before creatin a new round

	// check 1: the round has not already been created
	auction, err := s.QueryAuctionRound(ctx, auctionID, newRound)
	if auction != nil {
		return fmt.Errorf("Cannot create new round: round already exists")
	}

	// check 2: there was there a previous round
	previousRound := newRound - 1

	auction, err = s.QueryAuctionRound(ctx, auctionID, previousRound)
	if err != nil {
		return fmt.Errorf("Cannot create round until previous round is created")
	}

	// check 3: confirm that Demand >= Supply for the previous round before creating a new round
	if auction.Sold >= auction.Demand {
		return fmt.Errorf("Cannot create new round: demand is not yet greater than supply")
	}

	// If all three checks have passed, create a new round

	bidders := make(map[string]Bidder)

	auction.Round = newRound
	auction.Price = auction.Price + incrementAmount
	auction.Bidders = bidders
	auction.Demand = 0

	newAuctionRoundJSON, err := json.Marshal(auction)
	if err != nil {
		return err
	}
	// create a composite key for the new round
	newAuctionRoundKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(auction.Round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().PutState(newAuctionRoundKey, newAuctionRoundJSON)
	if err != nil {
		return fmt.Errorf("failed to create new auction round: %v", err)
	}

	// create an event to notify buyers and sellers of a new round
	err = ctx.GetStub().SetEvent("CreateNewRound", []byte(auctionID))
	if err != nil {
		return fmt.Errorf("event failed to register: %v", err)
	}

	return nil
}

// CloseAuctionRound closes a given round of the auction. This prevents
// bids from being added to the auction round, signaling that auction has
// reached a steady state.
func (s *SmartContract) CloseAuctionRound(ctx contractapi.TransactionContextInterface, auctionID string, round int) error {

	auction, err := s.QueryAuctionRound(ctx, auctionID, round)
	if err != nil {
		return fmt.Errorf("Error getting auction round from state")
	}

	status := auction.Status
	if status != "open" {
		return fmt.Errorf("Can only close an open auction")
	}

	// compelete series of checks before the auction can be closed.
	// checks confirms if the auction is still active before it can
	// be clossed
	err = s.closeAuctionChecks(ctx, auction)
	if err != nil {
		return fmt.Errorf("Cannot close round, round and auction is still active")
	}

	auction.Status = string("closed")

	closedAuction, _ := json.Marshal(auction)

	// create a composite key for the new round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// put the updated auction round in state
	err = ctx.GetStub().PutState(auctionKey, closedAuction)
	if err != nil {
		return fmt.Errorf("failed to close auction: %v", err)
	}

	// create an event that a round has closed
	err = ctx.GetStub().SetEvent("CloseRound", []byte(auctionID))
	if err != nil {
		return fmt.Errorf("event failed to register: %v", err)
	}

	return nil
}

// EndAuction defines the closed round as final stage of the auction.
// all other auction rounds are deleted from state.
func (s *SmartContract) EndAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {

	auction, err := s.QueryAuction(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("Error getting auction round from state")
	}

	// find if a round has been closed. If a round is closed, declare round final.
	closedRound := false
	for _, auctionRound := range auction {
		if auctionRound.Status == "closed" {
			closedRound = true
			auctionRound.Status = "final"
		}
	}

	// error if no round has been closed
	if closedRound == false {
		return fmt.Errorf("Cannot end auction. No rounds have been closed.")
	}

	// remove all open rounds
	for _, auctionRound := range auction {

		auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(auctionRound.Round)})
		if err != nil {
			return fmt.Errorf("failed to create composite key: %v", err)
		}

		if auctionRound.Status != "final" {
			err = ctx.GetStub().DelState(auctionKey)
			if err != nil {
				return fmt.Errorf("failed to delete auction round %v: %v", auctionKey, err)
			}
		}
	}

	// create an event that the auction has ended.
	err = ctx.GetStub().SetEvent("EndAuction", []byte(auctionID))
	if err != nil {
		return fmt.Errorf("event failed to register: %v", err)
	}

	return nil
}

//closeAuctionChecks completes a series of checks to see if the auction is still active before
// closing a round.
func (s *SmartContract) closeAuctionChecks(ctx contractapi.TransactionContextInterface, auction *AuctionRound) error {

	// check 1: check that all bids have been added to the round
	err := checkForHigherBid(ctx, auction.Price, auction.ItemSold, auction.Bidders)
	if err != nil {
		return fmt.Errorf("Cannot close auction: %v", err)
	}

	// check 2: check that all asks have been added to the round
	err = checkForLowerAsk(ctx, auction.Price, auction.ItemSold, auction.Sellers)
	if err != nil {
		return fmt.Errorf("Cannot close auction: %v", err)
	}

	// check 3: if supply is less than demand
	// check if there is another round. If there is, run the same checks
	// on that round
	if auction.Sold < auction.Demand {

		newRound := auction.Round + 1
		nextAuctionRound, err := s.QueryAuctionRound(ctx, auction.ID, newRound)
		if nextAuctionRound == nil {
			return fmt.Errorf("Need to start new round before this round can be closed")
		}

		if nextAuctionRound.Sold <= nextAuctionRound.Demand {
			err = s.closeAuctionChecks(ctx, nextAuctionRound)
			if err != nil {
				return fmt.Errorf("Next round is still active")
			} else {
				return fmt.Errorf("Cannot close non-final round")
			}
		}

	}

	return nil
}
