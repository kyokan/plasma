import Input from './Input';
import Output from './Output';
import {toBuffer} from '../util/numbers';
import * as ejs from 'ethereumjs-util';
import {keccak256} from '../crypto/hash';
import BN = require('bn.js');

/**
 * The body of a Plasma transaction. We draw a distinction between
 * the body of a transaction and its signatures to match the
 * transaction encoding expected by the Plasma smart contract. The
 * contract expects transactions to be RLP-encoded as follows:
 *
 * ```
 *  [[Blknum1, TxIndex1, Oindex1, DepositNonce1, Input1ConfirmSig,
 *    Blknum2, TxIndex2, Oindex2, DepositNonce2, Input2ConfirmSig,
 *    NewOwner, Denom1, NewOwner, Denom2, Fee],
 *   [Signature1, Signature2]]
 * ```
 *
 * The TransactionBody, in this case, is the first element of the above array.
 *
 * The `inputXConfirmSig` fields represent the confirm signatures belonging to the
 * transaction referenced by the `blockNum`, `txIdx`, and `outIdx` fields of the
 * referenced input - in other words, the 'previous' transaction in the UTXO DAG.
 * By committing to confirm signatures in this manner, anyone can use them to challenge
 * an invalid exit once an output is spent.
 */
export default class TransactionBody {
  /**
   * The first input to this transaction. For deposits, the input will have
   * a `blockNumber`, `txIdx`, and `outIdx` of zero and a `depositNonce` set
   * to the nonce of the deposit being spent.
   *
   * A transaction must always define `input0`.
   */
  public readonly input0: Input;

  /**
   * The second input to this transaction. This may be set to a 'zero'
   * input if there are no coins being spent via this input.
   */
  public readonly input1: Input;

  /**
   *
   */
  public readonly output0: Output;

  public readonly output1: Output;

  /**
   * The block number this transaction was included in. For transactions
   * that have not been included in a block yet, this will be zero.
   */
  public readonly blockNum: number;

  /**
   * The index of this transaction within the block it was included in. For
   * transactions that have not been included in a block yet, this will be
   * zero.
   */
  public readonly txIdx: number;

  /**
   * The confirm signature belonging to the transaction whose outputs are
   * referenced by `input0`.
   */
  public readonly input0ConfirmSig: Buffer;

  /**
   * The confirm signature belonging to the transaction whose outputs are
   * referenced by `input1`.
   */
  public readonly input1ConfirmSig: Buffer;

  /**
   * The fee paid to the root node to include this transaction.
   */
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
      ...this.input0.toArray(),
      this.input0ConfirmSig,
      ...this.input1.toArray(),
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