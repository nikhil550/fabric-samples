/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
*/

'use strict';

class AuctionRound {

  constructor(id, round, price, item, demand, quantity,sold) {
    this.id = id;
    this.round = round;
    this.price = price;
    this.item = item;
    this.demand = demand;
    this.quantity = quantity;
    this.sold = sold;
    this.joined = false;
  }

  join() {
    this.joined = true;
  }

  updateAuction(demand, quantity, sold) {
    this.demand = demand;
    this.quantity = quantity;
    this.sold = sold;
  }

}

module.exports = AuctionRound;
