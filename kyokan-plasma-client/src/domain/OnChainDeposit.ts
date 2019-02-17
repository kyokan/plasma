import BN = require('bn.js');

export interface OnChainDeposit {
  nonce: BN
  owner: string
  amount: BN
  createdAt: string
  ethBlockNum: string
}