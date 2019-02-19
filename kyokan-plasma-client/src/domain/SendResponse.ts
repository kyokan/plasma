import Transaction from './Transaction';

export interface SendResponse {
  transaction: Transaction
  inclusion: {
    merkleRoot: Buffer
    blockNumber: number
    transactionIndex: number
  }
}