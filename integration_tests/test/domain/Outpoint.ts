import BN = require('bn.js');
import {ConfirmedTransactionWire} from '../lib/PlasmaRPC';
import ConfirmedTransaction from './ConfirmedTransaction';

export default class Outpoint {
  public txIdx: number;

  public blockNum: number;

  public outIdx: number;

  public amount: BN;

  public transaction: ConfirmedTransaction;

  constructor (txIdx: number, blockNum: number, outIdx: number, amount: BN, transaction: ConfirmedTransaction) {
    this.txIdx = txIdx;
    this.blockNum = blockNum;
    this.outIdx = outIdx;
    this.amount = amount;
    this.transaction = transaction;
  }

  static fromWireTx (txWire: ConfirmedTransactionWire, owner: string): Outpoint {
    const tx = ConfirmedTransaction.fromConfirmedTransactionWire(txWire);
    const outIdx = owner === tx.output0.newOwner ? 0 : 1;

    return new Outpoint(
      tx.txIdx,
      tx.blockNum,
      outIdx,
      outIdx === 0 ? tx.output0.amount : tx.output1.amount,
      tx,
    );
  }
}