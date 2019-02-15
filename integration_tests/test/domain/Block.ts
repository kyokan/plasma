import {BlockWire} from '../lib/PlasmaRPC';
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

  static fromWire (blockWire: BlockWire): Block {
    return new Block(
      {
        merkleRoot: blockWire.block.header.merkleRoot,
        rlpMerkleRoot: blockWire.block.header.rlpMerkleRoot,
        prevHash: blockWire.block.header.prevHash,
        number: Number(blockWire.block.header.number),
      },
      blockWire.block.hash,
      blockWire.confirmedTransactions.map((ct) => ConfirmedTransaction.fromWire(ct)),
    );
  }
}