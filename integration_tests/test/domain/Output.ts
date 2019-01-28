import BN = require('bn.js');
import {parseHex, toHex} from '../lib/parseHex';
import {keccak256} from '../lib/hash';
import {fromBNWire, OutputWire, toBNWire} from '../lib/PlasmaRPC';
import {ZERO_ADDRESS} from './Addresses';
import {toBig, toBuffer} from '../lib/numbers';

export default class Output {
  public readonly newOwner: string;

  public readonly amount: BN;

  public readonly depositNonce: BN;

  constructor (newOwner: string, amount: BN, depositNonce: BN) {
    this.newOwner = newOwner;
    this.amount = amount;
    this.depositNonce = depositNonce;
  }

  public hash () {
    const buf = Buffer.concat([
      parseHex(this.newOwner),
      this.amount.toBuffer('be'),
      this.depositNonce.toBuffer('be'),
    ]);
    return keccak256(buf);
  }

  public toRPC (): OutputWire {
    return {
      newOwner: parseHex(this.newOwner),
      amount: toBNWire(this.amount),
      depositNonce: toBNWire(this.depositNonce),
    };
  }

  public toArray () {
    return [
      toBuffer(this.newOwner, 20),
      toBuffer(this.amount),
    ];
  }

  static zero () {
    return new Output(
      ZERO_ADDRESS,
      toBig(0),
      toBig(0),
    );
  }

  static fromWire (output: OutputWire): Output {
    return new Output(
      toHex(output.newOwner),
      fromBNWire(output.amount),
      fromBNWire(output.depositNonce),
    );
  }
}