import TransactionBuilder from '../domain/TransactionBuilder';
import {assert} from '../util/assert';
import ConfirmedTransaction from '../domain/ConfirmedTransaction';
import PlasmaContract from '../contract/PlasmaContract';
import {addressesEqual} from '../util/addresses';
import IRootClient from '../rpc/IRootClient';
import BN = require('bn.js');
import {Signer} from '../crypto/Signer';

/**
 * Executes a send on the sidechain.
 */
export default class SendOperation {
  private readonly client: IRootClient;

  private readonly contract: PlasmaContract;

  private readonly from: string;

  private value: BN | null = null;

  private to: string | null = null;

  private fee: BN | null = null;

  private depositNonce: BN | null = null;

  /**
   * Constructs a SendOperation.
   *
   * @param contract An instance of the Plasma smart contract.
   * @param client An instance of the Plasma root node gRPC client.
   * @param from The address whose outpoints are to be exited.
   */
  constructor (client: IRootClient, contract: PlasmaContract, from: string) {
    this.client = client;
    this.contract = contract;
    this.from = from;
  }

  /**
   * The address to send funds to.
   *
   * @param to The address.
   */
  public toAddress (to: string): SendOperation {
    this.to = to;
    return this;
  }

  /**
   * The amount of funds to send.
   *
   * @param value The amount.
   */
  public forValue (value: BN): SendOperation {
    this.value = value;
    return this;
  }

  /**
   * The fee amount to pay to the root node.
   *
   * @param fee The fee.
   */
  public withFee (fee: BN): SendOperation {
    this.fee = fee;
    return this;
  }

  /**
   * A deposit nonce. Set when spending a deposit.
   *
   * @param depositNonce
   */
  public withDepositNonce (depositNonce: BN): SendOperation {
    this.depositNonce = depositNonce;
    return this;
  }

  /**
   * Performs the send. Will throw an error if `to`, `value`, or
   * `fee` are not defined.
   *
   * Performing a send entails the following steps:
   *
   * 1. Looking up the address's UTXO set.
   * 2. Choosing the appropriate UTXOs to spend.
   * 3. Generating and signing a transaction.
   * 4. Sending that transaction to the root node.
   * 5. Consuming the returned inclusion receipt, generating confirm signatures,
   * and forwarding them to the root node.
   *
   * @param signer
   */
  public async send (signer: Signer): Promise<ConfirmedTransaction> {
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
    await tx.sign(signer);
    const confirmData = await this.client.send(tx);
    const inclusion = confirmData.inclusion;
    const confirmedTx = new ConfirmedTransaction(confirmData.transaction, null);
    await confirmedTx.confirmSign(signer, inclusion.merkleRoot);
    await this.client.confirm(confirmedTx);
    return confirmedTx;
  }
}
