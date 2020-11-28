
'use strict';

const { Gateway, Wallets } = require('fabric-network');
const path = require('path');
const { buildCCPOrg1, buildCCPOrg2, buildWallet } = require('../../../test-application/javascript/AppUtil.js');
const { buildCAClient, registerAndEnrollUser, enrollAdmin } = require('../../../test-application/javascript/CAUtil.js');


const myChannel = 'mychannel';
const myChaincodeName = 'auction';

const mspOrg1 = 'Org1MSP';
const mspOrg2 = 'Org2MSP';



async function main() {

  try {
    console.log('\n--> Enrolling the Org1 CA admin');
    const ccpOrg1 = buildCCPOrg1();
    const caOrg1Client = buildCAClient(FabricCAServices, ccpOrg1, 'ca.org1.example.com');

    const walletPathOrg1 = path.join(__dirname, 'wallet/org1');
    const walletOrg1 = await buildWallet(Wallets, walletPathOrg1);
    await enrollAdmin(caOrg2Client, walletOrg2, mspOrg2);

  } catch (error) {
      console.error(`Error enrolling Org1 admin: ${error}`);
      process.exit(1);
  }

  try {
    console.log('\n--> Enrolling the Org2 CA admin');
    const ccpOrg2 = buildCCPOrg2();
    const caOrg2Client = buildCAClient(FabricCAServices, ccpOrg2, 'ca.org2.example.com');

    const walletPathOrg2 = path.join(__dirname, 'wallet/org2');
    const walletOrg2 = await buildWallet(Wallets, walletPathOrg2);
    await enrollAdmin(caOrg2Client, walletOrg2, mspOrg2);

  } catch (error) {
    console.error(`Error enrolling Org1 admin: ${error}`);
    process.exit(1);
  }
}

main();
