import { assert } from 'chai';
import PlasmaClient from './lib/PlasmaClient';
import {assertBigEqual} from './lib/assertBigEqual';
import {toBig} from './lib/numbers';
import {wait} from './lib/wait';
import {GetBlockResponse} from './lib/PlasmaRPC';

describe('Basic functionality', () => {
  let client: PlasmaClient;

  before(() => {
    client = PlasmaClient.getShared();
  });

  it('should return a zero balance for a non-existent account', async () => {
    const res = await client.getBalance('0x0000000000000000000000000000000000000000');
    assertBigEqual(res, toBig(0));
  });

  it('should return the genesis block', async function () {
    this.timeout(30000);

    let res: GetBlockResponse|null = null;
    for (let i = 0; i < 3; i++) {
      try {
        res = await client.getBlock(1);
      } catch (e) {
        await wait(3000);
      }
    }

    if (!res) {
      throw new Error('Timed out waiting for genesis block.');
    }

    assert.equal(res.block.header.number, '1');
    assert.equal(res.block.hash.toString('hex'), '75af368e110be48098d9d7b4d6009d5c20323895dd426534500834ffd5cef62c');
  });
});