import ConfirmedTransaction from './ConfirmedTransaction';

/**
 * A Plasma block's header. The `merkleRoot` is committed to chain.
 */
export interface BlockHeader {
  merkleRoot: Buffer
  rlpMerkleRoot: Buffer
  prevHash: Buffer
  number: number
}

/**
 * A Plasma block.
 */
export default class Block {
  public header: BlockHeader;

  public hash: Buffer;

  public transactions: ConfirmedTransaction[];

  constructor (header: BlockHeader, hash: Buffer, transactions: ConfirmedTransaction[]) {
    this.header = header;
    this.hash = hash;
    this.transactions = transactions;
  }
}