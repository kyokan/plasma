import PlasmaClient from '../lib/PlasmaClient';
import TransactionBuilder from '../lib/TransactionBuilder';
import {assert} from 'chai';
import ConfirmedTransaction from './ConfirmedTransaction';
import PlasmaContract from '../lib/PlasmaContract';
import BN = require('bn.js');
import {addressesEqual} from '../util/addresses';
import Transaction from './Transaction';

export default class SendOperation {
  private readonly client: PlasmaClient;

  private readonly contract: PlasmaContract;

  private readonly from: string;

  private value: BN | null = null;

  private to: string | null = null;

  private fee: BN | null = null;

  private depositNonce: BN | null = null;

  constructor (client: PlasmaClient, contract: PlasmaContract, from: string) {
    this.client = client;
    this.contract = contract;
    this.from = from;
  }

  public toAddress (to: string): SendOperation {
    this.to = to;
    return this;
  }

  public forValue (value: BN): SendOperation {
    this.value = value;
    return this;
  }

  public withFee (fee: BN): SendOperation {
    this.fee = fee;
    return this;
  }

  public withDepositNonce (depositNonce: BN): SendOperation {
    this.depositNonce = depositNonce;
    return this;
  }

  public async send (privateKey: Buffer): Promise<void> {
    assert(this.to, 'a to address must be set');
    assert(this.value, 'a value must be set');
    assert(this.fee, 'a fee must be set');

    const builder = new TransactionBuilder(this.from)
      .forValue(this.value!)
      .toAddress(this.to!)
      .withFee(this.fee!);

    if (this.depositNonce) {
      const deposit = await this.contract.depositFor(this.depositNonce);
      if (!addressesEqual(this.from, deposit.owner)) {
        throw new Error('cannot spend non-owned deposit');
      }
      if (this.value!.gt(deposit.amount)) {
        throw new Error('cannot spend more than the deposit');
      }

      builder.withDepositNonce(this.depositNonce, deposit.amount);
    } else {
      const utxos = await this.client.getUTXOs(this.from);
      builder.withUTXOs(utxos);
    }

    const tx = builder.build();
    tx.sign(privateKey);
    const confirmData = await this.client.send(tx);
    const inclusion = confirmData.inclusion;
    const confirmedTx = new ConfirmedTransaction(Transaction.fromWire(confirmData.transaction), null)
    confirmedTx.confirmSign(privateKey, inclusion.merkleRoot);
    await this.client.confirm(confirmedTx);
  }
}