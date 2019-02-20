import {toBig} from 'kyokan-plasma-client/lib/util/numbers';
import IRootClient from 'kyokan-plasma-client/lib/rpc/IRootClient';
import {assertBigEqual} from './lib/assertBigEqual';
import bothClients from './lib/bothClients';

bothClients((client: IRootClient) => describe('Basic functionality', () => {
  it('should return a zero balance for a non-existent account', async () => {
    const res = await client.getBalance('0x0000000000000000000000000000000000000000');
    assertBigEqual(res, toBig(0));
  });
}));