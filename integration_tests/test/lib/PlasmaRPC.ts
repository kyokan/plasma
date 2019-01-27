import BN = require('bn.js');

export interface BNWire {
  hex: string
}

export function toBNWire (num: BN | number): BNWire {
  if (!(num instanceof BN)) {
    num = new BN(num);
  }

  return {
    hex: num.toString('hex'),
  };
}

export interface GetBalanceResponse {
  balance: BNWire
}

export interface GetBlockResponse {
  block: {
    header: {
      merkleRoot: Buffer
      rlpMerkleRoot: Buffer
      prevHash: Buffer
      number: string
    },
    hash: Buffer
  }
}

export interface InputWire {
  owner: Buffer
  depositNonce: BNWire
  blockNum: BNWire
  txIdx: BNWire
  outIdx: BNWire
}

export interface OutputWire {
  newOwner: Buffer
  amount: BNWire
  depositNonce: BNWire
}

export interface TransactionResponse {
  signatures: Buffer[]
  transaction: {
    input0: InputWire
    sig0: Buffer
    input1: InputWire
    sig1: Buffer
    output0: OutputWire
    fee: BNWire
    blockNum: BNWire
    txIdx: BNWire
  }
}

export interface GetOutputsResponse {
  confirmedTransactions: TransactionResponse[]
}