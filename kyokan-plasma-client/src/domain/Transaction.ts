import TransactionBody from './TransactionBody';
import {ethSign} from '../crypto/sign';
import ejs = require('ethereumjs-util');

export default class Transaction {
  public readonly body: TransactionBody;

  public signature0: Buffer | null = null;

  public signature1: Buffer | null = null;

  constructor (body: TransactionBody, signature0: Buffer | null, signature1: Buffer | null) {
    this.body = body;
    this.signature0 = signature0;
    this.signature1 = signature1;
  }

  toArray () {
    return [
      this.body.toArray(),
      [
        this.signature0,
        this.signature1,
      ],
    ];
  }

  toRLP (): Buffer {
    return (ejs as any).rlp.encode(this.toArray()) as Buffer;
  }

  sign (privateKey: Buffer) {
    const hash = this.body.sigHash();
    const sig = ethSign(hash, privateKey);
    this.signature0 = sig;
    this.signature1 = sig;
  }
}