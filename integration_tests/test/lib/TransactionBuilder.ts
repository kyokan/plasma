import { assert } from 'chai';
import {Outpoint} from '../domain/Outpoint';
import Transaction from '../domain/Transaction';
import BN = require('bn.js');
import {selectUTXOs} from './selectUTXOs';
import Input from '../domain/Input';
import {toBig} from './numbers';
import Output from '../domain/Output';
import Output from '../domain/Output';

export default class TransactionBuilder {
  private readonly utxos: Outpoint[];

  private readonly from: string;

  private to: string | null = null;

  private value: BN | null = null;

  private fee: BN | null = null;

  constructor (utxos: Outpoint[], from: string) {
    this.utxos = utxos;
    this.from = from;
  }

  public toAddress (to: string): TransactionBuilder {
    this.to = to;
    return this;
  }

  public forValue (value: BN): TransactionBuilder {
    this.value = value;
    return this;
  }

  public withFee (fee: BN): TransactionBuilder {
    this.fee = fee;
    return this;
  }

  public build (): Transaction {
    assert(this.to, 'a to address must be set');
    assert(this.value, 'a value must be set');
    assert(this.fee, 'a fee must be set');

    const totalAmount = this.value!.add(this.fee!);
    const outpoints = selectUTXOs(this.utxos, totalAmount);
    const totalOutpoints = outpoints.reduce((total: BN, outpoint: Outpoint) => {
      return total.add(outpoint.amount);
    }, toBig(0));

    const outputs = [];
    if (totalOutpoints.cmp(totalAmount) === 0) {
      outputs.push(new Output(
        this.to!,
        this.value!,
        toBig(0),
      ));
    } else {
      outputs.push(new Output(
        this.to!,
        this.value!,
        toBig(0),
      ), new Output(
        this.from,
        totalOutpoints.sub(totalAmount),
        toBig(0)
      ));
    }

    return new Transaction(
      new Input(
        outpoints[0].txIdx,
        outpoints[0].blockNum,
        outpoints[0].outIdx,
        this.from,
        toBig(0),
      ),
      outpoints[1] ? new Input(
        outpoints[0].txIdx,
        outpoints[0].blockNum,
        outpoints[0].outIdx,
        this.from,
        toBig(0),
      ) : Input.zero(),
      outputs[0],
      outputs[1] ? outputs[1] : Output.zero(),
      0,
      0,
      null,
      null,
      this.fee!
    );
  }
}