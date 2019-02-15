import PlasmaContract from './lib/PlasmaContract';
import { assert } from 'chai';
import {toBig} from './lib/numbers';
import {Config} from './Config';
import PlasmaClient from './lib/PlasmaClient';
import {withRetryCondition} from './lib/withRetries';
import SendOperation from './domain/SendOperation';
import BN = require('bn.js');

describe('Deposits', () => {
  let contract: PlasmaContract;
  let client: PlasmaClient;

  before(() => {
    contract = PlasmaContract.getShared();
    client = PlasmaClient.getShared();
  });

  it('should allow spends via the deposit nonce', async function () {
    this.timeout(60000);
    const depBal = toBig(1000);
    await contract.deposit(depBal, Config.USER_ADDRESSES[1]);
    const nonce = (await contract.depositNonce()).sub(new BN(1));
    const sendOp = new SendOperation(client, contract, Config.USER_ADDRESSES[1])
      .forValue(toBig(100))
      .toAddress(Config.USER_ADDRESSES[2])
      .withFee(toBig(1))
      .withDepositNonce(nonce);
    await sendOp.send(Config.PRIVATE_KEYS[1]);
    await withRetryCondition<BN>(() => client.getBalance(Config.USER_ADDRESSES[1]), (r) => r.eq(toBig(899)), 30);
    await withRetryCondition<BN>(() => client.getBalance(Config.USER_ADDRESSES[2]), (r) => r.eq(toBig(100)), 30);
  });

  it('should disallow subsequent spends from used deposit nonces', async function () {
    this.timeout(60000);

    const nonce = (await contract.depositNonce()).sub(new BN(1));
    const sendOp = new SendOperation(client, contract, Config.USER_ADDRESSES[1])
      .forValue(toBig(100))
      .toAddress(Config.USER_ADDRESSES[2])
      .withFee(toBig(1))
      .withDepositNonce(nonce);
    await assert.isRejected(sendOp.send(Config.PRIVATE_KEYS[1]));
  });
});