/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const path = require('path');

const { Gateway, Wallets } = require('fabric-network');

const { buildCCPOrg1, buildCCPOrg2, buildWallet } = require('../../../../test-application/javascript/AppUtil.js');

exports.initGatewayForOrg1 = async (user) => {

	const ccpOrg1 = buildCCPOrg1();
	const walletPathOrg1 = path.join(__dirname, '../wallet/org1');
	const walletOrg1 = await buildWallet(Wallets, walletPathOrg1);

	try {
		// Create a new gateway for connecting to Org's peer node.
		const gateway = new Gateway();
		//connect using Discovery enabled
		await gateway.connect(ccpOrg1,
			{ wallet: walletOrg1, identity: user, discovery: { enabled: true, asLocalhost: true } });

		return gateway;
	} catch (error) {
		console.error(`Error in connecting to gateway for Org1: ${error}`);
		process.exit(1);
	}
}

exports.initGatewayForOrg2 = async (user) => {

	const ccpOrg2 = buildCCPOrg2();
	const walletPathOrg2 = path.join(__dirname, '../wallet/org2');
	const walletOrg2 = await buildWallet(Wallets, walletPathOrg2);

	try {
		// Create a new gateway for connecting to Org's peer node.
		const gateway = new Gateway();
		await gateway.connect(ccpOrg2,
			{ wallet: walletOrg2, identity: user, discovery: { enabled: true, asLocalhost: true } });


		return gateway;
	} catch (error) {
		console.error(`Error in connecting to gateway for Org2: ${error}`);
		process.exit(1);
	}
}

exports.sleep = (ms) => {
	return new Promise((resolve) => setTimeout(resolve, ms));
}
