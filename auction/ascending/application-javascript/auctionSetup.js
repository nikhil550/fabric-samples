
'use strict';

const { Wallets } = require('fabric-network');
const FabricCAServices = require('fabric-ca-client');

const { buildCCPOrg1, buildCCPOrg2, buildWallet } = require('../../../test-application/javascript/AppUtil.js');
const { buildCAClient, registerAndEnrollUser, enrollAdmin } = require('../../../test-application/javascript/CAUtil.js');

const { initGatewayForOrg1, initGatewayForOrg2} = require('./lib/connect.js');
const { bid, ask } = require('./lib/bidAsk.js');
const { registerEnrollAuctionAdmin } = require('./lib/registerAdmin.js');

const path = require('path');

const mspOrg1 = 'Org1MSP';
const mspOrg2 = 'Org2MSP';

const item = 'tickets';

async function main() {

  // build connection information for org1
  const ccpOrg1 = buildCCPOrg1();
  const caOrg1Client = buildCAClient(FabricCAServices, ccpOrg1, 'ca.org1.example.com');

  const walletPathOrg1 = path.join(__dirname, 'wallet/org1');
  const walletOrg1 = await buildWallet(Wallets, walletPathOrg1);

  // build connection information for org2
  const ccpOrg2 = buildCCPOrg2();
  const caOrg2Client = buildCAClient(FabricCAServices, ccpOrg2, 'ca.org2.example.com');

  const walletPathOrg2 = path.join(__dirname, 'wallet/org2');
  const walletOrg2 = await buildWallet(Wallets, walletPathOrg2);

  // enrolling the Org1 CA admin
  try {
    console.log('\n--> Enrolling the Org1 CA admin');
    await enrollAdmin(caOrg1Client, walletOrg1, mspOrg1);

  } catch (error) {
      console.error(`Error enrolling Org1 admin: ${error}`);
      process.exit(1);
  }

  // enrolling the Org2 CA admin
  try {
    console.log('\n--> Enrolling the Org2 CA admin');
    await enrollAdmin(caOrg2Client, walletOrg2, mspOrg2);

  } catch (error) {
    console.error(`Error enrolling Org1 admin: ${error}`);
    process.exit(1);
  }

  // register and enrolling buyer and sellers from org1
  try {
    console.log('\n--> Enrolling buyers and sellers from Org1');
    var users = ["bidder1", "bidder2", "seller1", "seller2"];
    var userID;
    for (userID of users) {
      await registerAndEnrollUser(caOrg1Client, walletOrg1, mspOrg1, userID, 'org1.department1');
  }

  // register and enrolling auction admin from org1
  console.log('\n--> Enrolling auction admin from Org1');
  await registerEnrollAuctionAdmin(caOrg1Client, walletOrg1, mspOrg1, 'auctionAdmin', 'org1.department1');

  } catch (error) {
    console.error(`Error enrolling Org1 identities: ${error}`);
    process.exit(1);
  }

  // register and enrolling buyer and sellers from org2
  try {
    console.log('\n--> Enrolling buyers and sellers from Org2');
    var users = ["bidder3", "bidder4", "bidder5", "seller3", "seller4"];
    var userID;
    for (userID of users) {
      await registerAndEnrollUser(caOrg2Client, walletOrg2, mspOrg2 ,userID, 'org2.department1');
    }

  // register and enrolling auction admin from org2
  console.log('\n--> Enrolling auction admin from Org2');
  await registerEnrollAuctionAdmin(caOrg2Client, walletOrg2, mspOrg2,'auctionAdmin', 'org2.department1');

  } catch (error) {
    console.error(`Error enrolling Org2 identities: ${error}`);
    process.exit(1);
  }

  // submit bids for the
  try {

    console.log('\n--> submitting bids and asks from Org1');

    const gatewayBidder1 = await initGatewayForOrg1("bidder1")
    await bid(gatewayBidder1,"Org1MSP",item,"20","20");
    gatewayBidder1.disconnect();

    const gatewayBidder2 = await initGatewayForOrg1("bidder2")
    await bid(gatewayBidder2,"Org1MSP",item,"20","40");
    gatewayBidder2.disconnect();

    const gatewaySeller1 = await initGatewayForOrg1("seller1")
    await ask(gatewaySeller1,"Org1MSP",item,"20","30");
    gatewaySeller1.disconnect();

    const gatewaySeller2 = await initGatewayForOrg1("seller2")
    await ask(gatewaySeller2,"Org1MSP",item,"20","50");
    gatewaySeller2.disconnect();

        console.log('\n--> submitting bids and asks from Org2');

        const gatewayBidder3 = await initGatewayForOrg2("bidder3")
        await bid(gatewayBidder3,"Org2MSP",item,"20","60");
        gatewayBidder3.disconnect();

        const gatewayBidder4 = await initGatewayForOrg2("bidder4")
        await bid(gatewayBidder4,"Org2MSP",item,"20","80");
        gatewayBidder4.disconnect();

        const gatewayBidder5 = await initGatewayForOrg2("bidder5")
        await bid(gatewayBidder5,"Org2MSP",item,"20","100");
        gatewayBidder5.disconnect();

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
