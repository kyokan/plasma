import BN = require('bn.js');
import ConfirmedTransaction from './ConfirmedTransaction';

export default class Outpoint {
  public txIdx: number;

  public blockNum: number;

  public outIdx: number;

  public amount: BN;

  public confirmSig: Buffer;

  public transaction: ConfirmedTransaction | null;

  constructor (blockNum: number, txIdx: number, outIdx: number, amount: BN, confirmSig: Buffer, transaction: ConfirmedTransaction | null) {
    this.blockNum = blockNum;
    this.txIdx = txIdx;
    this.outIdx = outIdx;
    this.amount = amount;
    this.confirmSig = confirmSig;
    this.transaction = transaction;
  }
}