import PlasmaClient from './lib/PlasmaClient';
import {assertBigEqual} from './lib/assertBigEqual';
import {toBig} from './lib/numbers';

describe('Basic functionality', () => {
  let client: PlasmaClient;

  before(() => {
    client = PlasmaClient.getShared();
  });

  it('should return a zero balance for a non-existent account', async () => {
    const res = await client.getBalance('0x0000000000000000000000000000000000000000');
    assertBigEqual(res, toBig(0));
  });
});