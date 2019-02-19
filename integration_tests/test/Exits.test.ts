import PlasmaContract from 'kyokan-plasma-client/lib/contract/PlasmaContract';
import IRootClient from 'kyokan-plasma-client/lib/rpc/IRootClient';
import SendOperation from 'kyokan-plasma-client/lib/operations/SendOperation';
import ExitOperation from 'kyokan-plasma-client/lib/operations/ExitOperation';
import {toBig} from 'kyokan-plasma-client/lib/util/numbers';
import {Config} from './Config';
import {withRetryCondition} from './lib/withRetries';
import {wait} from './lib/wait';
import {EventLog} from 'web3/types';
import {assert} from 'chai';
import SharedRootClient from './lib/SharedRootClient';
import SharedContract from './lib/SharedContract';
import BN = require('bn.js');

describe('Exits', () => {
  let contract: PlasmaContract;
  let client: IRootClient;
  let depositNonce: BN;

  before(async function () {
    this.timeout(60000);
    contract = SharedContract.get();
    client = SharedRootClient.get();
    const depBal = toBig(10000000000);
    await contract.deposit(depBal, Config.USER_ADDRESSES[4]);
    depositNonce = (await contract.depositNonce()).sub(toBig(1));
    const sendBal = toBig(999000000);
    const sendOp = new SendOperation(client, contract, Config.USER_ADDRESSES[4])
      .forValue(sendBal)
      .toAddress(Config.USER_ADDRESSES[5])
      .withFee(toBig(1))
      .withDepositNonce(depositNonce);
    await sendOp.send(Config.PRIVATE_KEYS[4]);
    await withRetryCondition(() => client.getBalance(Config.USER_ADDRESSES[5]), (r) => r.eq(sendBal), 30);
  });

  it('should exit', async function () {
    this.timeout(30000);
    await wait(5000);
    const outpoints = await client.getUTXOs(Config.USER_ADDRESSES[5]);
    const exitOp = new ExitOperation(contract, client, Config.USER_ADDRESSES[5])
      .withOutpoint(outpoints[0])
      .withCommittedFee(toBig(500000));

    await exitOp.exit();
  });

  it('should challenge exits', async function () {
    this.timeout(60000);

    // send to party A
    const sendBal = toBig(1000000);
    const sendOp = new SendOperation(client, contract, Config.USER_ADDRESSES[4])
      .forValue(sendBal)
      .toAddress(Config.USER_ADDRESSES[6])
      .withFee(toBig(1));
    await sendOp.send(Config.PRIVATE_KEYS[4]);
    await withRetryCondition(() => client.getBalance(Config.USER_ADDRESSES[6]), (r) => r.eq(sendBal), 30);
    // record outpoints after A receives
    const outpoints = await client.getUTXOs(Config.USER_ADDRESSES[6]);

    // A sends to B
    console.log('Spending input...');
    const sendOtherBal = toBig(800000);
    const sendOpOther = new SendOperation(client, contract, Config.USER_ADDRESSES[6])
      .forValue(sendOtherBal)
      .toAddress(Config.USER_ADDRESSES[7])
      .withFee(toBig(1));
    await sendOpOther.send(Config.PRIVATE_KEYS[6]);
    await withRetryCondition(() => client.getBalance(Config.USER_ADDRESSES[7]), (r) => r.eq(sendOtherBal), 30);
    // wait for block submission
    await wait(20000);

    const outpoint = outpoints[0];
    // A attempts to exit a spent input
    const exitOp = new ExitOperation(contract, client, Config.USER_ADDRESSES[6])
      .withOutpoint(outpoint)
      .withCommittedFee(toBig(500000));

    console.log('Attempting exit with spent input...');
    await exitOp.exit();

    // let chainsaw catch the exit
    await wait(5000);

    // query the exit's status
    const exitData = await contract.challengedExits();
    const outExit = exitData.find((e: EventLog) => {
      const retVals = e.returnValues;
      return retVals.position[0] === outpoint.blockNum.toString() &&
        retVals.position[1] === outpoint.txIdx.toString() &&
        retVals.position[2] === outpoint.outIdx.toString() &&
        retVals.position[3] === '0' &&
        retVals.owner.toLowerCase() === Config.USER_ADDRESSES[6].toLowerCase() &&
        retVals.amount === '500000';
    });
    assert.isDefined(outExit, 'no exit challenge found');
  });

  it('should challenge deposit exits', async function () {
    await contract.startDepositExit(depositNonce, toBig(500000), Config.USER_ADDRESSES[4]);
    await wait(8000);

    const exitData = await contract.challengedExits();
    const outExit = exitData.find((e: EventLog) => {
      const retVals = e.returnValues;
      return retVals.position[0] === '0' &&
        retVals.position[1] === '0' &&
        retVals.position[2] === '0' &&
        retVals.position[3] === depositNonce.toString(10) &&
        retVals.owner.toLowerCase() === Config.USER_ADDRESSES[4].toLowerCase() &&
        retVals.amount === '9999500000';
    });
    assert.isDefined(outExit, 'no exit challenge found');
  });
});