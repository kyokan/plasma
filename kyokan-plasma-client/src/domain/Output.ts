import BN = require('bn.js');
import {parseHex} from '../util/hex';
import {keccak256} from '../crypto/hash';
import {ZERO_ADDRESS} from './Addresses';
import {toBig, toBuffer} from '../util/numbers';

/**
 * A transaction output.
 */
export default class Output {
  /**
   * The new owner of this output.
   */
  public readonly owner: string;

  /**
   * The value of this output.
   */
  public readonly amount: BN;

  constructor (owner: string, amount: BN) {
    this.owner = owner;
    this.amount = amount;
  }

  /**
   * Returns the hash of this output for use in signatures.
   */
  public hash () {
    const buf = Buffer.concat([
      parseHex(this.owner),
      this.amount.toBuffer('be'),
    ]);
    return keccak256(buf);
  }

  /**
   * Serializes the output to an array that can be encoded as RLP.
   */
  public toArray () {
    return [
      toBuffer(this.owner, 20),
      toBuffer(this.amount),
    ];
  }

  /**
   * Returns a 'zero' output - that is, an output whose
   * fields are all zero. Used in [[TransactionBody]] objects
   * to represent null outputs.
   */
  static zero () {
    return new Output(
      ZERO_ADDRESS,
      toBig(0),
    );
  }
}