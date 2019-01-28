import Input from './Input';
import Output from './Output';
import {toBNWire} from '../lib/PlasmaRPC';
import {toBuffer} from '../lib/numbers';
import * as ejs from 'ethereumjs-util';
import {keccak256} from '../lib/hash';
import BN = require('bn.js');

export default class Transaction {
  public readonly input0: Input;

  public readonly input1: Input;

  public readonly output0: Output;

  public readonly output1: Output;

  public readonly blockNum: number;

  public readonly txIdx: number;

  public sig0: Buffer;

  public sig1: Buffer;

  public readonly fee: BN;

  constructor (input0: Input, input1: Input, output0: Output, output1: Output, blockNum: number, txIdx: number, sig0: Buffer | null, sig1: Buffer | null, fee: BN) {
    this.input0 = input0;
    this.input1 = input1;
    this.output0 = output0;
    this.output1 = output1;
    this.blockNum = blockNum;
    this.txIdx = txIdx;
    this.sig0 = sig0 || Buffer.from('');
    this.sig1 = sig1 || Buffer.from('');
    this.fee = fee;
  }

  toRPC (signatures: Buffer[]) {
    return {
      transaction: {
        blockNum: toBNWire(this.blockNum),
        txIdx: toBNWire(this.txIdx),
        sig0: this.sig0,
        sig1: this.sig1,
        fee: toBNWire(this.fee),
        input0: this.input0.toRPC(),
        input1: this.input1.toRPC(),
        output0: this.output0.toRPC(),
        output1: this.output1.toRPC(),
      },
      signatures,
    };
  }

  toArray () {
    return [
      ...this.input0.toArray(),
      this.sig0,
      ...this.input1.toArray(),
      this.sig1,
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
    console.log('rlp is', rlp.toString('hex'));
    return keccak256(rlp);
  }

  sign (privateKey: Buffer): [Buffer, Buffer] {
    this.sig0 = this.input0.sign(privateKey);
    this.sig1 = this.input1.sign(privateKey);
    const confSigHash = this.sigHash();
    console.log('sigHashConf', confSigHash.toString('hex'));
    const sig = ejs.ecsign(confSigHash, privateKey);
    const confSig = Buffer.concat([sig.r, sig.s, Buffer.from([sig.v])]);
    return [confSig, confSig];
  }
}