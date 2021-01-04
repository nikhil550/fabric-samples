/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
*/

'use strict';

class AuctionRound {

  constructor(id, round, item, demand, quantity) {
    this.id = id;
    this.round = round;
    this.item = item;
    this.demand = demand;
    this.quantity = quantity;
  }
}

module.exports = AuctionRound;
