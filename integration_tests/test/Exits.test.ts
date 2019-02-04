import PlasmaContract from './lib/PlasmaContract';
import PlasmaClient from './lib/PlasmaClient';
import {toBig} from './lib/numbers';
import {Config} from './Config';
import {withRetryCondition} from './lib/withRetries';
import SendOperation from './domain/SendOperation';
import ExitOperation from './domain/ExitOperation';
import {wait} from './lib/wait';

describe('Exits', () => {
  let contract: PlasmaContract;
  let client: PlasmaClient;

  before(async function () {
    this.timeout(60000);
    contract = PlasmaContract.getShared();
    client = PlasmaClient.getShared();
    const depBal = toBig(1000000000);
    await contract.deposit(depBal, Config.USER_ADDRESSES[4]);
    await withRetryCondition(() => client.getBalance(Config.USER_ADDRESSES[4]), (r) => r.eq(depBal), 30);
    const sendBal = toBig(999000000);
    const sendOp = new SendOperation(client, Config.USER_ADDRESSES[4])
      .forValue(sendBal)
      .toAddress(Config.USER_ADDRESSES[5])
      .withFee(toBig(1));
    await sendOp.send(Config.PRIVATE_KEYS[4]);
    await withRetryCondition(() => client.getBalance(Config.USER_ADDRESSES[5]), (r) => r.eq(sendBal), 30);
  });

  it('should exit', async function () {
    this.timeout(30000);
    await wait(15000);
    const outpoints = await client.getUTXOs(Config.USER_ADDRESSES[5]);
    const exitOp = new ExitOperation(contract, client, Config.USER_ADDRESSES[5])
      .withOutpoint(outpoints[0])
      .withCommittedFee(toBig(500000));

    // NOTE: Need to use PRIVATE_KEYS[4] below to generate the
    // correct confirm sigs.
    await exitOp.exit(Config.PRIVATE_KEYS[4]);
  });
});