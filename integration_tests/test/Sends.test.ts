import PlasmaContract from './lib/PlasmaContract';
import PlasmaClient from './lib/PlasmaClient';
import {toBig} from './lib/numbers';
import {Config} from './Config';
import {withRetryCondition} from './lib/withRetries';
import {assertBigEqual} from './lib/assertBigEqual';
import {wait} from './lib/wait';
import SendOperation from './domain/SendOperation';

describe('Sends', () => {
  let contract: PlasmaContract;
  let client: PlasmaClient;

  before(async function () {
    this.timeout(60000);
    contract = PlasmaContract.getShared();
    client = PlasmaClient.getShared();
    const depBal = toBig(1000);
    await contract.deposit(depBal, Config.USER_ADDRESSES[2]);
    await withRetryCondition(() => client.getBalance(Config.USER_ADDRESSES[2]), (r) => r.eq(depBal), 30);
  });

  it('should debit the sender and credit the receiver when a send is initiated', async () => {
    const sent = toBig(532);
    const fee = toBig(1);
    const sendOp = new SendOperation(client, Config.USER_ADDRESSES[2])
      .forValue(sent)
      .toAddress(Config.USER_ADDRESSES[3])
      .withFee(fee);
    await sendOp.send(Config.PRIVATE_KEYS[2]);
    await wait(10000);
    assertBigEqual(await client.getBalance(Config.USER_ADDRESSES[3]), sent);
    assertBigEqual(await client.getBalance(Config.USER_ADDRESSES[2]), toBig(1000).sub(sent).sub(fee));
  });
});