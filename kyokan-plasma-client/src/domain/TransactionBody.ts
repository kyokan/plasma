import Input from './Input';
import Output from './Output';
import {toBuffer} from '../util/numbers';
import * as ejs from 'ethereumjs-util';
import {keccak256} from '../crypto/hash';
import BN = require('bn.js');

export default class TransactionBody {
  public readonly input0: Input;

  public readonly input1: Input;

  public readonly output0: Output;

  public readonly output1: Output;

  public readonly blockNum: number;

  public readonly txIdx: number;

  public readonly input0ConfirmSig: Buffer;

  public readonly input1ConfirmSig: Buffer;

  public readonly fee: BN;

  constructor (input0: Input, input1: Input, output0: Output, output1: Output, blockNum: number, txIdx: number, input0ConfirmSig: Buffer, input1ConfirmSig: Buffer, fee: BN) {
    this.input0 = input0;
    this.input1 = input1;
    this.output0 = output0;
    this.output1 = output1;
    this.blockNum = blockNum;
    this.txIdx = txIdx;
    this.input0ConfirmSig = input0ConfirmSig;
    this.input1ConfirmSig = input1ConfirmSig;
    this.fee = fee;
  }

  toArray () {
    return [
      ...this.input0.toConfirmSigArray(),
      this.input0ConfirmSig,
      ...this.input1.toConfirmSigArray(),
      this.input1ConfirmSig,
      ...this.output0.toArray(),
      ...this.output1.toArray(),
      toBuffer(this.fee),
    ];
  }

  toRLP (): Buffer {
    return (ejs as any).rlp.encode(this.toArray()) as Buffer;
  }

  sigHash () {
    const rlp = this.toRLP();
    return keccak256(rlp);
  }
}