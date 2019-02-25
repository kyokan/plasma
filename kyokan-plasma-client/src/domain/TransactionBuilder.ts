import {assert} from '../util/assert';
import Outpoint from './Outpoint';
import Transaction from './Transaction';
import {selectUTXOs} from './selectUTXOs';
import Input from './Input';
import {toBig} from '../util/numbers';
import Output from './Output';
import TransactionBody from './TransactionBody';
import BN = require('bn.js');

export default class TransactionBuilder {
  private utxos: Outpoint[] | null = null;

  private readonly from: string;

  private to: string | null = null;

  private value: BN | null = null;

  private fee: BN | null = null;

  private depositNonce: BN | null = null;

  private depositAmount: BN | null = null;

  constructor (from: string) {
    this.from = from;
  }

  public withUTXOs (utxos: Outpoint[]): TransactionBuilder {
    this.utxos = utxos;
    return this;
  }

  public withDepositNonce (depositNonce: BN, depositAmount: BN): TransactionBuilder {
    this.depositNonce = depositNonce;
    this.depositAmount = depositAmount;
    return this;
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
    let outpoints: Outpoint[];
    if (this.depositNonce) {
      outpoints = [new Outpoint(
        0,
        0,
        0,
        this.depositAmount!,
        Buffer.alloc(65),
        null,
      )];
    } else if (this.utxos) {
      outpoints = selectUTXOs(this.utxos, totalAmount);
    } else {
      throw new Error('must provide either deposit nonce or UTXOs');
    }

    const totalOutpoints = outpoints.reduce((total: BN, outpoint: Outpoint) => {
      return total.add(outpoint.amount);
    }, toBig(0));

    const outputs = [];
    if (totalOutpoints.cmp(totalAmount) === 0) {
      outputs.push(new Output(
        this.to!,
        this.value!,
      ));
    } else {
      outputs.push(new Output(
        this.to!,
        this.value!,
      ), new Output(
        this.from,
        totalOutpoints.sub(totalAmount),
      ));
    }

    const body = new TransactionBody(
      new Input(
        outpoints[0].blockNum,
        outpoints[0].txIdx,
        outpoints[0].outIdx,
        this.depositNonce ? this.depositNonce : toBig(0),
      ),
      outpoints[1] ? new Input(
        outpoints[1].blockNum,
        outpoints[1].txIdx,
        outpoints[1].outIdx,
        toBig(0),
      ) : Input.zero(),
      outputs[0],
      outputs[1] ? outputs[1] : Output.zero(),
      0,
      0,
      outpoints[0].confirmSig,
      outpoints[1] ? outpoints[1].confirmSig : Buffer.alloc(65),
      this.fee!,
    );

    return new Transaction(
      body,
      null,
      null,
    );
  }
}