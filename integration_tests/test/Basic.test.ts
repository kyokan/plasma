import {assert} from 'chai';
import PlasmaClient from './lib/PlasmaClient';
import {assertBigEqual} from './lib/assertBigEqual';
import {toBig} from './lib/numbers';
import {wait} from './lib/wait';
import Block from './domain/Block';

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

    let res: Block | null = null;
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

    assert.equal(res.header.number, 1);
    assert.equal(res.hash.toString('hex'), '3ba210900ba59a2539595d8c091cf3eaaf7c968d24124144763c5d683abde0b1');
  });
});