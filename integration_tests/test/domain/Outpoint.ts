import BN = require('bn.js');
import {ConfirmedTransactionWire} from '../lib/PlasmaRPC';
import ConfirmedTransaction from './ConfirmedTransaction';

export default class Outpoint {
  public txIdx: number;

  public blockNum: number;

  public outIdx: number;

  public amount: BN;

  public confirmSig: Buffer;

  public transaction: ConfirmedTransaction|null;

  constructor (txIdx: number, blockNum: number, outIdx: number, amount: BN, confirmSig: Buffer, transaction: ConfirmedTransaction|null) {
    this.txIdx = txIdx;
    this.blockNum = blockNum;
    this.outIdx = outIdx;
    this.amount = amount;
    this.confirmSig = confirmSig;
    this.transaction = transaction;
  }

  static fromWireTx (txWire: ConfirmedTransactionWire, owner: string): Outpoint {
    const tx = ConfirmedTransaction.fromWire(txWire);
    const body = tx.transaction.body;
    const outIdx = owner === body.output0.owner ? 0 : 1;

    if (!tx.confirmSignatures) {
      throw new Error('cannot create outpoint from unconfirmed transaction');
    }

    return new Outpoint(
      body.txIdx,
      body.blockNum,
      outIdx,
      outIdx === 0 ? body.output0.amount : body.output1.amount,
      tx.confirmSignatures[outIdx],
      tx,
    );
  }
}