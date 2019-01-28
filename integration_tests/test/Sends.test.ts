import PlasmaContract from './lib/PlasmaContract';
import PlasmaClient from './lib/PlasmaClient';
import {toBig} from './lib/numbers';
import {Config} from './Config';
import {withRetryCondition} from './lib/withRetries';
import TransactionBuilder from './lib/TransactionBuilder';
import {assertBigEqual} from './lib/assertBigEqual';
import {wait} from './lib/wait';

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
    const utxos = await client.getUTXOs(Config.USER_ADDRESSES[2]);
    const sent = toBig(532);
    const fee = toBig(1);
    const tx = new TransactionBuilder(utxos, Config.USER_ADDRESSES[2])
      .forValue(sent)
      .toAddress(Config.USER_ADDRESSES[3])
      .withFee(fee)
      .build();

    const confirmSigs = tx.sign(Config.PRIVATE_KEYS[2]);
    await client.send(tx, confirmSigs);
    await wait(10000);
    assertBigEqual(await client.getBalance(Config.USER_ADDRESSES[3]), sent);
    assertBigEqual(await client.getBalance(Config.USER_ADDRESSES[2]), toBig(1000).sub(sent).sub(fee));
  });
});