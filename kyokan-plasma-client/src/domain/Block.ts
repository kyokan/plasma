import ConfirmedTransaction from './ConfirmedTransaction';

/**
 * A Plasma block's header. The `merkleRoot` is committed to chain.
 */
export interface BlockHeader {
  merkleRoot: Buffer
  number: number
}

/**
 * A Plasma block.
 */
export default class Block {
  public header: BlockHeader;

  public transactions: ConfirmedTransaction[];

  constructor (header: BlockHeader, transactions: ConfirmedTransaction[]) {
    this.header = header;
    this.transactions = transactions;
  }
}