import Transaction from './Transaction';
import * as ejs from 'ethereumjs-util';
import {ConfirmedTransactionWire} from '../lib/PlasmaRPC';
import {sha256} from '../lib/hash';
import {ethSign} from '../lib/sign';

export default class ConfirmedTransaction {
  public readonly transaction: Transaction;

  public confirmSignatures: [Buffer, Buffer] | null;

  constructor (transaction: Transaction, confirmSignatures: [Buffer, Buffer] | null) {
    this.transaction = transaction;
    this.confirmSignatures = confirmSignatures;
  }

  confirmHash (merkleRoot: Buffer) {
    return sha256(Buffer.concat([
      sha256(this.transaction.toRLP()),
      merkleRoot,
    ]));
  }

  confirmSign (privateKey: Buffer, merkleRoot: Buffer) {
    const confirmSigHash = this.confirmHash(merkleRoot);
    const confirmSig = ethSign(confirmSigHash, privateKey);
    this.confirmSignatures = [
      confirmSig, confirmSig,
    ];
  }

  static fromWire (wireTx: ConfirmedTransactionWire): ConfirmedTransaction {
    return new ConfirmedTransaction(
      Transaction.fromWire(wireTx.transaction),
      [wireTx.confirmSig0, wireTx.confirmSig1],
    );
  }
}