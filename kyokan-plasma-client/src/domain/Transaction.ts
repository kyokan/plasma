import TransactionBody from './TransactionBody';
import {ethSign} from '../crypto/sign';
import ejs = require('ethereumjs-util');
import {Signer} from '../crypto/Signer';

/**
 * A Plasma transaction.
 */
export default class Transaction {
  /**
   * The body of the transaction.
   */
  public readonly body: TransactionBody;

  /**
   * The unlock signature authorizing `body.input0`.
   */
  public signature0: Buffer | null = null;

  /**
   * The unlock signature authorizing `body.input1`.
   */
  public signature1: Buffer | null = null;

  constructor (body: TransactionBody, signature0: Buffer | null, signature1: Buffer | null) {
    this.body = body;
    this.signature0 = signature0;
    this.signature1 = signature1;
  }

  /**
   * Serializes the transaction to an array that can be encoded as RLP.
   */
  toArray () {
    return [
      this.body.toArray(),
      [
        this.signature0,
        this.signature1,
      ],
    ];
  }

  /**
   * Serializes the transaction to an RLP-encoded `Buffer`.
   */
  toRLP (): Buffer {
    return (ejs as any).rlp.encode(this.toArray()) as Buffer;
  }

  /**
   * Generates the authorization signatures for both inputs.
   *
   * @param signer
   */
  async sign (signer: Signer) {
    const hash = this.body.sigHash();
    const sig = await signer.ethSign(hash);
    this.signature0 = sig;
    this.signature1 = sig;
  }
}
