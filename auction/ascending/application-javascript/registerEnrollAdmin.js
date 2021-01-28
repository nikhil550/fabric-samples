/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Wallets } = require('fabric-network');
const FabricCAServices = require('fabric-ca-client');
const path = require('path');
const { buildCAClient } = require('../../../test-application/javascript/CAUtil.js');
const { buildCCPOrg1, buildCCPOrg2, buildWallet } = require('../../../test-application/javascript/AppUtil.js');

const mspOrg1 = 'Org1MSP';
const mspOrg2 = 'Org2MSP';
const adminUserId = 'admin';
const adminUserPasswd = 'adminpw';

async function registerAndEnrollUser(caClient, wallet, orgMspId, userId, affiliation) {
	try {
		// Check to see if we've already enrolled the user
		const userIdentity = await wallet.get(userId);
		if (userIdentity) {
			console.log(`An identity for the user ${userId} already exists in the wallet`);
			return;
		}

		// Must use an admin to register a new user
		const adminIdentity = await wallet.get(adminUserId);
		if (!adminIdentity) {
			console.log('An identity for the admin user does not exist in the wallet');
			console.log('Enroll the admin user before retrying');
			return;
		}

		// build a user object for authenticating with the CA
		const provider = wallet.getProviderRegistry().getProvider(adminIdentity.type);
		const adminUser = await provider.getUserContext(adminIdentity, adminUserId);
		// Register the user, enroll the user, and import the new identity into the wallet.
		// if affiliation is specified by client, the affiliation value must be configured in CA
		const secret = await caClient.register({
			affiliation: affiliation,
			enrollmentID: userId,
			role: 'client',
      attrs: [{name: "role", value: "auctionAdmin", ecert: true }],
		}, adminUser);
		const enrollment = await caClient.enroll({
			enrollmentID: userId,
			enrollmentSecret: secret
		});
		const x509Identity = {
			credentials: {
				certificate: enrollment.certificate,
				privateKey: enrollment.key.toBytes(),
			},
			mspId: orgMspId,
			type: 'X.509',
		};
		await wallet.put(userId, x509Identity);
		console.log(`Successfully registered and enrolled user ${userId} and imported it into the wallet`);
	} catch (error) {
		console.error(`Failed to register user : ${error}`);
	}
};

async function connectToOrg1CA(UserID) {
    console.log('\n--> Register and enrolling new user');
    const ccpOrg1 = buildCCPOrg1();
    const caOrg1Client = buildCAClient(FabricCAServices, ccpOrg1, 'ca.org1.example.com');

    const walletPathOrg1 = path.join(__dirname, 'wallet/org1');
    const walletOrg1 = await buildWallet(Wallets, walletPathOrg1);

    await registerAndEnrollUser(caOrg1Client, walletOrg1, mspOrg1, UserID, 'org1.department1');

}

async function connectToOrg2CA(UserID) {
    console.log('\n--> Register and enrolling new user');
    const ccpOrg2 = buildCCPOrg2();
    const caOrg2Client = buildCAClient(FabricCAServices, ccpOrg2, 'ca.org2.example.com');

    const walletPathOrg2 = path.join(__dirname, 'wallet/org2');
    const walletOrg2 = await buildWallet(Wallets, walletPathOrg2);

    await registerAndEnrollUser(caOrg2Client, walletOrg2, mspOrg2, UserID, 'org2.department1');

}
async function main() {

    if (process.argv[2] == undefined && process.argv[3] == undefined) {
        console.log("Usage: node registerEnrollUser.js org userID");
        process.exit(1);
    }

    const org = process.argv[2];
    const userId = process.argv[3];

    try {

      if (org == 'Org1' || org == 'org1') {
        await connectToOrg1CA(userId);
      }
      else if (org == 'Org2' || org == 'org2') {
        await connectToOrg2CA(userId);
      } else {
        console.log("Usage: node registerEnrollUser.js org userID");
        console.log("Org must be Org1 or Org2");
      }
    } catch (error) {
        console.error(`Error in enrolling admin: ${error}`);
        process.exit(1);
    }
}

main();
