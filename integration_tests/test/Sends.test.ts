import PlasmaContract from './lib/PlasmaContract';
import PlasmaClient from './lib/PlasmaClient';
import {toBig} from './lib/numbers';
import {Config} from './Config';
import {withRetryCondition} from './lib/withRetries';

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

  it('should debit the sender and credit the receiver when a send is initiated', async  () => {
    console.log(JSON.stringify(await client.getUTXOs(Config.USER_ADDRESSES[2])));
  });
});