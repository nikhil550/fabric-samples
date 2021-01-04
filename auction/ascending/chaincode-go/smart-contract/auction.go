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
	ID       string            `json:"id"`
	Round    int               `json:"round"`
	Status   string            `json:"status"`
	ItemSold string            `json:"item"`
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
}

const incrementAmount = 10

// CreateAuction creates on auction on the public channel. Each auction round is
// stored as a seperate key in the world state
func (s *SmartContract) CreateAuction(ctx contractapi.TransactionContextInterface, auctionID string, itemSold string, reservePrice int) error {

	existingAuction, err := s.QueryAuction(ctx, auctionID)
	if existingAuction != nil {
		return fmt.Errorf("Cannot create new auction: auction already exists")
	}

	sellers := make(map[string]Seller)
	bidders := make(map[string]Bidder)

	// Check 1: check if there is an ask from your org that is lower than the reserve price
	err = queryAllAsks(ctx, reservePrice, itemSold, sellers)
	if err != nil {
		return fmt.Errorf("seller has lower ask, cannot open a new auction at this price: %v", err)
	}

	// Create auction

	auction := Auction{
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

	// create a composite key using the round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(auction.Round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// put auction into state
	err = ctx.GetStub().PutState(auctionKey, auctionJSON)
	if err != nil {
		return fmt.Errorf("failed to put auction in public data: %v", err)
	}

	// create an event to notify buyers and sellers of a new auction
	err = ctx.GetStub().SetEvent("CreateAuction", auctionJSON)
	if err != nil {
		return fmt.Errorf("event failed to register: %v", err)
	}

	return nil
}

// CreateNewRound creates another round of the auction, which is a seperate key
// in world state. Bidders and sellers have the abiltiy to join the round at the
// new price
func (s *SmartContract) CreateNewRound(ctx contractapi.TransactionContextInterface, auctionID string, newRound int) error {

	// check 1: there was no previous round

	auction, err := s.QueryAuctionRound(ctx, auctionID, newRound)
	if auction != nil {
		return fmt.Errorf("Cannot create new round: round already exists")
	}

	// check 2: was there a previous round

	previousRound := newRound - 1

	auction, err = s.QueryAuctionRound(ctx, auctionID, previousRound)
	if err != nil {
		return fmt.Errorf("Cannot create round until previous round is created")
	}

	// check 3: confirm that Demand > Supply for the previous round before creating a new round

	if auction.Quantity >= auction.Demand {
		return fmt.Errorf("Cannot create new round: demand is not yet greater than supply")
	}

	// Because both checks passed, create a new round

	sellers := make(map[string]Seller)
	bidders := make(map[string]Bidder)

	auction.Round = newRound
	auction.Price = auction.Price + incrementAmount
	auction.Sellers = sellers
	auction.Bidders = bidders
	auction.Quantity = 0
	auction.Demand = 0

	newAuctionRoundJSON, err := json.Marshal(auction)
	if err != nil {
		return err
	}
	// create a composite key using the transaction ID
	newAuctionRoundKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(auction.Round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().PutState(newAuctionRoundKey, newAuctionRoundJSON)
	if err != nil {
		return fmt.Errorf("failed to close auction: %v", err)
	}

	// create an event to notify buyers and sellers of a new round
	err = ctx.GetStub().SetEvent("CreateNewRound", newAuctionRoundJSON)
	if err != nil {
		return fmt.Errorf("event failed to register: %v", err)
	}

	return nil
}

// CloseAuction can be used by the seller to close the auction. This prevents
// bids from being added to the auction, and allows users to reveal their bid
func (s *SmartContract) CloseAuctionRound(ctx contractapi.TransactionContextInterface, auctionID string, round int) error {

	auction, err := s.QueryAuctionRound(ctx, auctionID, round)
	if err != nil {
		return err
	}

	Status := auction.Status
	if Status != "open" {
		return fmt.Errorf("Can only close an open auction")
	}

	err = s.closeAuctionChecks(ctx, auction)
	if err != nil {
		return fmt.Errorf("Cannot closer round, round and auction is still active")
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

	err = ctx.GetStub().SetEvent("closeRound", closedAuction)
	if err != nil {
		return fmt.Errorf("event failed to register: %v", err)
	}

	return nil
}

// EndAuction closes all of the auction rounds for bidding. The closed round
// has status final. All other rounds are removed from state.
func (s *SmartContract) EndAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {

	auction, err := s.QueryAuction(ctx, auctionID)
	if err != nil {
		return err
	}

	// Find if round is closed. If a round is closed, declare found final.
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

	err = ctx.GetStub().SetEvent("EndAuction", []byte(auctionID))
	if err != nil {
		return fmt.Errorf("event failed to register: %v", err)
	}

	return nil
}

func (s *SmartContract) closeAuctionChecks(ctx contractapi.TransactionContextInterface, auction *Auction) error {

	// check 1: check that all bids have been added to the round

	err := queryAllBids(ctx, auction.Price, auction.ItemSold, auction.Bidders)
	if err != nil {
		return fmt.Errorf("Cannot close auction: %v", err)
	}

	// check 2: check that all asks have been added to the round

	err = queryAllAsks(ctx, auction.Price, auction.ItemSold, auction.Sellers)
	if err != nil {
		return fmt.Errorf("Cannot close auction: %v", err)
	}

	// check 3: D for the previous round before creating a new round

	if auction.Quantity <= auction.Demand {

		newRound := auction.Round + 1
		nextAuctionRound, err := s.QueryAuctionRound(ctx, auction.ID, newRound)
		if nextAuctionRound == nil {
			return fmt.Errorf("Need to start new round before this round can be closed")
		}

		err = s.closeAuctionChecks(ctx, nextAuctionRound)
		if err != nil {
			return fmt.Errorf("Next round is still active")
		}

	}

	return nil
}
