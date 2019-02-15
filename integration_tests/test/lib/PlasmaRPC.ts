import BN = require('bn.js');
import {toBig} from './numbers';

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

export function fromBNWire (num: BNWire): BN {
  return toBig(num.hex);
}

export interface GetBalanceResponse {
  balance: BNWire
}

export interface BlockWire {
  block: {
    header: {
      merkleRoot: Buffer
      rlpMerkleRoot: Buffer
      prevHash: Buffer
      number: string
    },
    hash: Buffer
  },
  confirmedTransactions: ConfirmedTransactionWire[]
}

export interface InputWire {
  depositNonce: BNWire
  blockNum: string
  txIdx: number
  outIdx: number
}

export interface OutputWire {
  owner: Buffer
  amount: BNWire
}

export interface TransactionBodyWire {
  input0: InputWire
  input0ConfirmSig: Buffer
  input1: InputWire
  input1ConfirmSig: Buffer
  output0: OutputWire
  output1: OutputWire
  fee: BNWire
  blockNum: string
  txIdx: number
}

export interface TransactionWire {
  body: TransactionBodyWire
  sig0: Buffer
  sig1: Buffer
}

export interface ConfirmedTransactionWire {
  confirmSig0: Buffer
  confirmSig1: Buffer
  transaction: TransactionWire
}

export interface GetOutputsResponse {
  confirmedTransactions: ConfirmedTransactionWire[]
}

export interface SendResponse {
  transaction: TransactionWire
  inclusion: {
    merkleRoot: Buffer
    blockNumber: number
    transactionIndex: number
  }
}

export interface GetConfirmationsResponse {
  authSig0: Buffer
  authSig1: Buffer
}

export interface ConfirmRequest {
  blockNumber: number
  transactionIndex: number
  confirmSig0: Buffer
  confirmSig1: Buffer
}