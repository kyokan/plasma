import PlasmaContract from 'kyokan-plasma-client/lib/contract/PlasmaContract';
import IRootClient from 'kyokan-plasma-client/lib/rpc/IRootClient';
import SendOperation from 'kyokan-plasma-client/lib/operations/SendOperation';
import {toBig} from 'kyokan-plasma-client/lib/util/numbers';
import {assert} from 'chai';
import {Config} from './Config';
import {withRetryCondition} from './lib/withRetries';
import SharedContract from './lib/SharedContract';
import bothClients from './lib/bothClients';
import {signerFromStr} from './util/signer';
import BN = require('bn.js');

bothClients((client: IRootClient) => describe('Deposits', () => {
  let contract: PlasmaContract;

  before(() => {
    contract = SharedContract.get();
  });

  it('should allow spends via the deposit nonce', async function () {
    this.timeout(60000);
    const depBal = toBig(1000);
    await contract.deposit(depBal, Config.USER_ADDRESSES[1]);
    const nonce = (await contract.depositNonce()).sub(new BN(1));
    const startBal1 = await client.getBalance(Config.USER_ADDRESSES[1]);
    const startBal2 = await client.getBalance(Config.USER_ADDRESSES[2]);
    const sendOp = new SendOperation(client, contract, Config.USER_ADDRESSES[1])
      .forValue(toBig(100))
      .toAddress(Config.USER_ADDRESSES[2])
      .withFee(toBig(1))
      .withDepositNonce(nonce);
    await sendOp.send(signerFromStr(Config.PRIVATE_KEYS[1]));
    await withRetryCondition<BN>(() => client.getBalance(Config.USER_ADDRESSES[1]), (r) => r.eq(startBal1.add(toBig(899))), 30);
    await withRetryCondition<BN>(() => client.getBalance(Config.USER_ADDRESSES[2]), (r) => r.eq(startBal2.add(toBig(100))), 30);
  });

  it('should disallow subsequent spends from used deposit nonces', async function () {
    this.timeout(60000);

    const nonce = (await contract.depositNonce()).sub(new BN(1));
    const sendOp = new SendOperation(client, contract, Config.USER_ADDRESSES[1])
      .forValue(toBig(100))
      .toAddress(Config.USER_ADDRESSES[2])
      .withFee(toBig(1))
      .withDepositNonce(nonce);
    await assert.isRejected(sendOp.send(signerFromStr(Config.PRIVATE_KEYS[1])));
  });
}));
