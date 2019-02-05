import Transaction from './Transaction';
import Input from './Input';
import Output from './Output';
import BN = require('bn.js');
import * as ejs from 'ethereumjs-util';
import {ConfirmedTransactionWire} from '../lib/PlasmaRPC';
import {sha256, tmSHA256} from '../lib/hash';
import {ethSign} from '../lib/sign';

export default class ConfirmedTransaction extends Transaction {
  public readonly confirmSignatures: [Buffer, Buffer];

  constructor (input0: Input, input1: Input, output0: Output, output1: Output, blockNum: number, txIdx: number, sig0: Buffer | null, sig1: Buffer | null, fee: BN, confirmSignatures: [Buffer, Buffer]) {
    super(input0, input1, output0, output1, blockNum, txIdx, sig0, sig1, fee);
    this.confirmSignatures = confirmSignatures;
  }

  toArray (): Buffer[] {
    throw new Error('not implemented directly');
  }

  toRLP (): Buffer {
    const arr = [
      super.toArray(),
      this.confirmSignatures,
    ];

    return (ejs as any).rlp.encode(arr) as Buffer;
  }

  authSign(privateKey: Buffer, merkleRoot: Buffer): [Buffer, Buffer] {
    const authSigHash = sha256(Buffer.concat([ sha256(this.toRLP()), merkleRoot ]));
    const authSig = ethSign(authSigHash, privateKey);
    return [authSig, authSig];
  }

  static fromTransaction (tx: Transaction, confirmSignatures: [Buffer, Buffer]) {
    return new ConfirmedTransaction(
      tx.input0,
      tx.input1,
      tx.output0,
      tx.output1,
      tx.blockNum,
      tx.txIdx,
      tx.sig0,
      tx.sig1,
      tx.fee,
      confirmSignatures,
    );
  }

  static fromConfirmedTransactionWire(wireTx: ConfirmedTransactionWire): ConfirmedTransaction {
    const tx = Transaction.fromWire(wireTx.transaction);
    return ConfirmedTransaction.fromTransaction(tx, wireTx.signatures as [Buffer, Buffer]);
  }
}