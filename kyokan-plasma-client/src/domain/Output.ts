import BN = require('bn.js');
import {parseHex} from '../util/hex';
import {keccak256} from '../crypto/hash';
import {ZERO_ADDRESS} from './Addresses';
import {toBig, toBuffer} from '../util/numbers';

export default class Output {
  public readonly owner: string;

  public readonly amount: BN;

  constructor (owner: string, amount: BN) {
    this.owner = owner;
    this.amount = amount;
  }

  public hash () {
    const buf = Buffer.concat([
      parseHex(this.owner),
      this.amount.toBuffer('be'),
    ]);
    return keccak256(buf);
  }

  public toArray () {
    return [
      toBuffer(this.owner, 20),
      toBuffer(this.amount),
    ];
  }

  static zero () {
    return new Output(
      ZERO_ADDRESS,
      toBig(0),
    );
  }
}