import {Uint64BE} from 'int64-buffer';
import {keccak256} from '../crypto/hash';
import {toBig, toBuffer} from '../util/numbers';
import * as ejs from 'ethereumjs-util';
import BN = require('bn.js');

/**
 * A transaction input.
 */
export default class Input {
  /**
   * The block number of the output being spent by this input.
   */
  public readonly blockNum: number;

  /**
   * The index of the transaction containing the output being spent by this input.
   */
  public readonly txIdx: number;

  /**
   * The index of the output being spent by this input.
   */
  public readonly outIdx: number;

  /**
   * The deposit nonce being spent by this input.
   */
  public readonly depositNonce: BN;

  /**
   * Constructs a new Input. Will throw an error if `depositNonce` is set alongside
   * any non-zero `blockNum`, `txIdx`, or `outIdx`.
   *
   * @param blockNum
   * @param txIdx
   * @param outIdx
   * @param depositNonce
   */
  constructor (blockNum: number, txIdx: number, outIdx: number, depositNonce?: BN) {
    if (depositNonce && (blockNum !== 0 || txIdx !== 0 || outIdx !== 0)) {
      throw new Error('cannot set a deposit nonce alongside blockNum, txIdx, or outIdx');
    }

    this.blockNum = blockNum;
    this.txIdx = txIdx;
    this.outIdx = outIdx;
    this.depositNonce = depositNonce || toBig(0);
  }

  /**
   * Returns the keccak256 hash of this input for use in signatures.
   */
  public hash (): Buffer {
    const buf = Buffer.concat([
      new Uint64BE(this.blockNum).toBuffer(),
      Buffer.from(`00000000${this.txIdx.toString(16)}`, 'hex'),
      ejs.toBuffer(this.outIdx),
    ]);
    return keccak256(buf);
  }

  /**
   * Serializes the input to an array that can be encoded as RLP.
   */
  public toArray () {
    return [
      toBuffer(this.blockNum),
      toBuffer(this.txIdx),
      toBuffer(this.outIdx),
      toBuffer(this.depositNonce),
    ];
  }

  /**
   * Serializes the input to an RLP-encoded `Buffer`.
   */
  public toRLP () {
    return (ejs as any).rlp.encode(this.toArray());
  }

  /**
   * Returns a 'zero' input - that is, an input whose fields are
   * all zero. Used in [[TransactionBody]] objects to represent
   * null inputs.
   */
  static zero (): Input {
    return new Input(
      0,
      0,
      0,
      toBig(0),
    );
  }
}