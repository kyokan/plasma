import Transaction from './Transaction';
import {sha256} from '../crypto/hash';
import {ethSign} from '../crypto/sign';

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
}