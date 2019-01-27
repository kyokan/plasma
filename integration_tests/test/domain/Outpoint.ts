import BN = require('bn.js');

export interface Outpoint {
  amount: BN
  txIdx: number
  blockNum: number
  outIdx: number
}