/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { prettyJSONString} = require('../../../test-application/javascript/AppUtil.js');
const { initGatewayForOrg1, initGatewayForOrg2} = require('./lib/connect.js');
const { bid, ask } = require('./lib/bidAsk.js');

const myChannel = 'mychannel';
const myChaincodeName = 'auction';
const item = 'good';

async function main() {
    try {

      const gatewayBidder1 = await initGatewayForOrg1("bidder1")
      await bid(gatewayBidder1,"Org1MSP",item,"20","20");
      gatewayBidder1.disconnect();

      const gatewayBidder2 = await initGatewayForOrg1("bidder2")
      await bid(gatewayBidder2,"Org1MSP",item,"20","40");
      gatewayBidder2.disconnect();

      const gatewayBidder3 = await initGatewayForOrg2("bidder3")
      await bid(gatewayBidder3,"Org2MSP",item,"20","60");
      gatewayBidder3.disconnect();

      const gatewayBidder4 = await initGatewayForOrg2("bidder4")
      await bid(gatewayBidder4,"Org2MSP",item,"20","80");
      gatewayBidder4.disconnect();

      const gatewayBidder5 = await initGatewayForOrg2("bidder5")
      await bid(gatewayBidder5,"Org2MSP",item,"20","100");
      gatewayBidder5.disconnect();

      const gatewaySeller1 = await initGatewayForOrg1("seller1")
      await ask(gatewaySeller1,"Org1MSP",item,"20","30");
      gatewaySeller1.disconnect();

      const gatewaySeller2 = await initGatewayForOrg1("seller2")
      await ask(gatewaySeller2,"Org1MSP",item,"20","50");
      gatewaySeller2.disconnect();

      const gatewaySeller3 = await initGatewayForOrg2("seller3")
      await ask(gatewaySeller3,"Org2MSP",item,"20","70");
      gatewaySeller3.disconnect();

      const gatewaySeller4 = await initGatewayForOrg2("seller4")
      await ask(gatewaySeller4,"Org2MSP",item,"20","90");
      gatewaySeller4.disconnect();

    } catch (error) {
		console.error(`******** FAILED to run the application: ${error}`);
    if (error.stack) {
        console.error(error.stack);
    }
    process.exit(1);
    }
}


main();
