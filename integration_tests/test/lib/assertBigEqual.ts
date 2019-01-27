import { assert } from 'chai';
import BN = require('bn.js');

export function assertBigEqual(a: BN, b: BN) {
  assert(a.eq(b), `${a.toString(10)} is not equal to ${b.toString(10)}`)
}