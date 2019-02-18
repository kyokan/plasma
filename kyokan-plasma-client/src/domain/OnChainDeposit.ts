import BN = require('bn.js');

/**
 * Represents the result of an on-chain deposit operation. The
 * returned `nonce` can be used to spend the deposit.
 */
export interface OnChainDeposit {
  nonce: BN
  owner: string
  amount: BN
  createdAt: string
  ethBlockNum: string
}