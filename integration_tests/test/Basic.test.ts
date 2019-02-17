import {toBig} from 'kyokan-plasma-client/lib/util/numbers';
import IRootClient from 'kyokan-plasma-client/lib/rpc/IRootClient';
import {assertBigEqual} from './lib/assertBigEqual';
import SharedRootClient from './lib/SharedRootClient';

describe('Basic functionality', () => {
  let client: IRootClient;

  before(() => {
    client = SharedRootClient.get();
  });

  it('should return a zero balance for a non-existent account', async () => {
    const res = await client.getBalance('0x0000000000000000000000000000000000000000');
    assertBigEqual(res, toBig(0));
  });
});