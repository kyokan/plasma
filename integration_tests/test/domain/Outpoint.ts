import BN = require('bn.js');
import {fromBNWire, TransactionWire} from '../lib/PlasmaRPC';
import {toHex} from '../lib/parseHex';

export interface Outpoint {
  amount: BN
  txIdx: number
  blockNum: number
  outIdx: number
}

export function fromWireTransaction (tx: TransactionWire, owner: string): Outpoint {
  const txIdx = fromBNWire(tx.txIdx).toNumber();
  const blockNum = fromBNWire(tx.blockNum).toNumber();
  const outIdx = owner === toHex(tx.output0.newOwner) ? 0 : 1;

  console.log(owner, toHex(tx.output0.newOwner));

  return {
    txIdx,
    blockNum,
    outIdx,
    amount: outIdx === 0 ? fromBNWire(tx.output0.amount) : fromBNWire(tx.output1.amount),
  };
}