
'use strict';

const { Wallets } = require('fabric-network');
const FabricCAServices = require('fabric-ca-client');
const path = require('path');
const { buildCCPOrg1, buildCCPOrg2, buildWallet } = require('../../../test-application/javascript/AppUtil.js');
const { buildCAClient, registerAndEnrollUser, enrollAdmin } = require('../../../test-application/javascript/CAUtil.js');
const { registerEnrollAuctionAdmin } = require('./lib/registerAdmin.js');

const mspOrg1 = 'Org1MSP';
const mspOrg2 = 'Org2MSP';


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

  // register and enrolling buyer and sellers from org2
  try {
    console.log('\n--> Enrolling buyers and sellers from Org1');
    var users = ["bidder1", "bidder2", "seller1", "seller2"];
    var userID;
    for (userID of users) {
      await registerAndEnrollUser(caOrg1Client, walletOrg1, mspOrg1, userID, 'org1.department1');
  }
  } catch (error) {
    console.error(`Error enrolling Org1 identities: ${error}`);
    process.exit(1);
  }

  // register and enrolling buyer and sellers from org2
  try {
    console.log('\n--> Enrolling buyers and sellers from Org1');
    var users = ["bidder3", "bidder4", "bidder5", "seller3", "seller4"];
    var userID;
    for (userID of users) {
      await registerAndEnrollUser(caOrg1Client, walletOrg1, mspOrg1,userID, 'org1.department1');
    }

    await registerEnrollAuctionAdmin(caOrg1Client, walletOrg1, mspOrg1, 'auctionAdmin', 'org1.department1');

  } catch (error) {
    console.error(`Error enrolling Org2 identities: ${error}`);
    process.exit(1);
  }

}

main();
