import PlasmaContract from './lib/PlasmaContract';
import {toBig} from './lib/numbers';
import {Config} from './Config';
import PlasmaClient from './lib/PlasmaClient';
import {assertBigEqual} from './lib/assertBigEqual';
import {withRetryCondition} from './lib/withRetries';
import BN = require('bn.js');

describe('Deposits', () => {
  let contract: PlasmaContract;
  let client: PlasmaClient;

  before(() => {
    contract = PlasmaContract.getShared();
    client = PlasmaClient.getShared();
  });

  it('should increase the user\'s balance upon deposit', async function () {
    this.timeout(60000);
    const depBal = toBig(1000);
    await contract.deposit(depBal, Config.USER_ADDRESSES[1]);
    const balance = await withRetryCondition<BN>(() => client.getBalance(Config.USER_ADDRESSES[1]), (r) => r.gt(toBig(0)), 30);
    assertBigEqual(balance, depBal);
  });

  it('should increase the user\'s balance upon further deposits', async function () {
    this.timeout(60000);
    const depBal = toBig(1500);
    await contract.deposit(depBal, Config.USER_ADDRESSES[1]);
    const balance = await withRetryCondition<BN>(() => client.getBalance(Config.USER_ADDRESSES[1]), (r) => r.gt(toBig(1000)), 30);
    assertBigEqual(balance, depBal.add(toBig(1000)));
  });
});