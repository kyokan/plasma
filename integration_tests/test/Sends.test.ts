import PlasmaContract from 'kyokan-plasma-client/lib/contract/PlasmaContract';
import IRootClient from 'kyokan-plasma-client/lib/rpc/IRootClient';
import SendOperation from 'kyokan-plasma-client/lib/operations/SendOperation';
import SharedContract from './lib/SharedContract';
import {toBig} from 'kyokan-plasma-client/lib/util/numbers';
import {Config} from './Config';
import {withRetryCondition} from './lib/withRetries';
import {assertBigEqual} from './lib/assertBigEqual';
import {wait} from './lib/wait';
import bothClients from './lib/bothClients';

bothClients((client: IRootClient) => describe('Sends', () => {
  let contract: PlasmaContract;

  before(async function () {
    this.timeout(60000);
    contract = SharedContract.get();
    const depBal = toBig(1000);
    await contract.deposit(depBal, Config.USER_ADDRESSES[2]);
    const nonce = (await contract.depositNonce()).sub(toBig(1));
    const sendAmount = toBig(990);
    const startBal3 = await client.getBalance(Config.USER_ADDRESSES[3]);
    const sendOp = new SendOperation(client, contract, Config.USER_ADDRESSES[2])
      .forValue(sendAmount)
      .toAddress(Config.USER_ADDRESSES[3])
      .withFee(toBig(1))
      .withDepositNonce(nonce);
    await sendOp.send(Config.PRIVATE_KEYS[2]);
    await withRetryCondition(() => client.getBalance(Config.USER_ADDRESSES[3]), (r) => r.eq(startBal3.add(sendAmount)), 30);
  });

  it('should debit the sender and credit the receiver when a send is initiated', async () => {
    const startBal3 = await client.getBalance(Config.USER_ADDRESSES[3]);
    const startBal4 = await client.getBalance(Config.USER_ADDRESSES[4]);
    const sendAmount = toBig(100);
    const sendOp = new SendOperation(client, contract, Config.USER_ADDRESSES[3])
      .forValue(sendAmount)
      .toAddress(Config.USER_ADDRESSES[4])
      .withFee(toBig(1));
    await sendOp.send(Config.PRIVATE_KEYS[3]);
    await wait(1000);
    assertBigEqual(await client.getBalance(Config.USER_ADDRESSES[3]), startBal3.sub(sendAmount).sub(toBig(1)));
    assertBigEqual(await client.getBalance(Config.USER_ADDRESSES[4]), startBal4.add(sendAmount));
  });
}));