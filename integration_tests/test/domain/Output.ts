import BN = require('bn.js');
import {parseHex, toHex} from '../lib/parseHex';
import {keccak256} from '../lib/hash';
import {fromBNWire, OutputWire, toBNWire} from '../lib/PlasmaRPC';
import {ZERO_ADDRESS} from './Addresses';
import {toBig, toBuffer} from '../lib/numbers';

export default class Output {
  public readonly owner: string;

  public readonly amount: BN;

  constructor (owner: string, amount: BN) {
    this.owner = owner;
    this.amount = amount;
  }

  public hash () {
    const buf = Buffer.concat([
      parseHex(this.owner),
      this.amount.toBuffer('be')
    ]);
    return keccak256(buf);
  }

  public toRPC (): OutputWire {
    return {
      owner: parseHex(this.owner),
      amount: toBNWire(this.amount),
    };
  }

  public toArray () {
    return [
      toBuffer(this.owner, 20),
      toBuffer(this.amount),
    ];
  }

  static zero () {
    return new Output(
      ZERO_ADDRESS,
      toBig(0),
    );
  }

  static fromWire (output: OutputWire): Output {
    return new Output(
      toHex(output.owner),
      fromBNWire(output.amount),
    );
  }
}