import PlasmaClient from '../lib/PlasmaClient';
import TransactionBuilder from '../lib/TransactionBuilder';
import {assert} from 'chai';
import BN = require('bn.js');
import ConfirmedTransaction from './ConfirmedTransaction';

export default class SendOperation {
  private readonly client: PlasmaClient;

  private readonly from: string;

  private value: BN | null = null;

  private to: string | null = null;

  private fee: BN | null = null;

  constructor (client: PlasmaClient, from: string) {
    this.client = client;
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

  public async send (privateKey: Buffer): Promise<void> {
    assert(this.to, 'a to address must be set');
    assert(this.value, 'a value must be set');
    assert(this.fee, 'a fee must be set');

    const utxos = await this.client.getUTXOs(this.from);
    const tx = new TransactionBuilder(utxos, this.from)
      .forValue(this.value!)
      .toAddress(this.to!)
      .withFee(this.fee!)
      .build();

    const confirmSigs = tx.sign(privateKey);
    const confirmData = await this.client.send(tx, confirmSigs);
    const confirmedTx = ConfirmedTransaction.fromTransaction(tx, confirmSigs);
    const authSigs = confirmedTx.authSign(privateKey, confirmData.merkleRoot);
    await this.client.confirm(confirmData.blockNumber, confirmData.transactionIndex, authSigs);
  }
}