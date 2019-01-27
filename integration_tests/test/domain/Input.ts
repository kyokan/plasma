import {Uint64BE} from 'int64-buffer';
import {keccak256} from '../lib/hash';
import {InputWire, toBNWire} from '../lib/PlasmaRPC';
import {parseHex} from '../lib/parseHex';
import {toBig, toBuffer} from '../lib/numbers';
import * as ejs from 'ethereumjs-util';
import {ZERO_ADDRESS} from './Addresses';
import BN = require('bn.js');

export default class Input {
  private readonly blockNum: number;

  private readonly txIdx: number;

  private readonly outIdx: number;

  private readonly owner: string;

  private readonly depositNonce: BN;

  constructor (blockNum: number, txIdx: number, outIdx: number, owner: string, depositNonce: BN) {
    this.blockNum = blockNum;
    this.txIdx = txIdx;
    this.outIdx = outIdx;
    this.owner = owner;
    this.depositNonce = depositNonce;
  }

  public hash (): Buffer {
    const buf = Buffer.concat([
      new Uint64BE(this.blockNum).toBuffer(),
      Buffer.from(`00000000${this.txIdx.toString(16)}`, 'hex'),
      ejs.toBuffer(this.outIdx),
    ]);
    return keccak256(buf);
  }

  public sigHash (): Buffer {
    const rlp = this.toRLP();
    return keccak256(rlp);
  }

  public sign(privateKey: Buffer): Buffer {
    const hash = this.sigHash();
    const sig = ejs.ecsign(hash, privateKey);
    return Buffer.concat([ sig.r, sig.s, Buffer.from([ sig.v ]) ]);
  }

  public toRPC (): InputWire {
    return {
      blockNum: toBNWire(this.blockNum),
      txIdx: toBNWire(this.txIdx),
      outIdx: toBNWire(this.outIdx),
      owner: parseHex(this.owner),
      depositNonce: toBNWire(this.depositNonce),
    };
  }

  public toArray () {
    return [
      toBuffer(this.blockNum),
      toBuffer(this.txIdx),
      toBuffer(this.outIdx),
      toBuffer(this.depositNonce),
      toBuffer(this.owner, 20),
    ];
  }

  public toRLP () {
    return (ejs as any).rlp.encode(this.toArray());
  }

  static zero (): Input {
    return new Input(
      0,
      0,
      0,
      ZERO_ADDRESS,
      toBig(0),
    );
  }
}