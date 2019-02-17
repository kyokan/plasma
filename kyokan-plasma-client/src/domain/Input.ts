import {Uint64BE} from 'int64-buffer';
import {keccak256} from '../crypto/hash';
import {toBig, toBuffer} from '../util/numbers';
import * as ejs from 'ethereumjs-util';
import BN = require('bn.js');

export default class Input {
  public readonly blockNum: number;

  public readonly txIdx: number;

  public readonly outIdx: number;

  public readonly depositNonce: BN;

  constructor (blockNum: number, txIdx: number, outIdx: number, depositNonce: BN) {
    this.blockNum = blockNum;
    this.txIdx = txIdx;
    this.outIdx = outIdx;
    this.depositNonce = depositNonce;
  }

  public hash (): Buffer {
    const buf = Buffer.concat([
      new Uint64BE(this.blockNum).toBuffer(),
      Buffer.from(`00000000${this.txIdx.toString(16)}`, 'hex'),
      ejs.toBuffer(this.outIdx),
    ]);
    return keccak256(buf);
  }

  public toArray () {
    return [
      toBuffer(this.blockNum),
      toBuffer(this.txIdx),
      toBuffer(this.outIdx),
      toBuffer(this.depositNonce),
    ];
  }

  public toConfirmSigArray () {
    return [
      toBuffer(this.blockNum),
      toBuffer(this.txIdx),
      toBuffer(this.outIdx),
      toBuffer(this.depositNonce),
    ];
  }

  public toRLP () {
    return (ejs as any).rlp.encode(this.toArray());
  }

  static zero (): Input {
    return new Input(
      0,
      0,
      0,
      toBig(0),
    );
  }
}