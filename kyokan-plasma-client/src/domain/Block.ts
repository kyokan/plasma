import ConfirmedTransaction from './ConfirmedTransaction';

export interface BlockHeader {
  merkleRoot: Buffer
  rlpMerkleRoot: Buffer
  prevHash: Buffer
  number: number
}

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