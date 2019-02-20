import {assert} from 'chai';
import Outpoint from '../domain/Outpoint';
import PlasmaContract from '../contract/PlasmaContract';
import MerkleTree from '../MerkleTree';
import {sha256} from '../crypto/hash';
import IRootClient from '../rpc/IRootClient';
import BN = require('bn.js');

/**
 * Constructs and executes an exit from the root chain.
 */
export default class ExitOperation {
  private contract: PlasmaContract;

  private client: IRootClient;

  private readonly from: string;

  private outpoint: Outpoint | null = null;

  private committedFee: BN | null = null;

  /**
   * Constructs an ExitOperation.
   *
   * @param contract An instance of the Plasma smart contract.
   * @param client An instance of the Plasma root node gRPC client.
   * @param from The address whose outpoints are to be exited.
   */
  constructor (contract: PlasmaContract, client: IRootClient, from: string) {
    this.contract = contract;
    this.client = client;
    this.from = from;
  }

  /**
   * Sets the outpoint to be exited.
   *
   * @param outpoint The outpoint.
   */
  withOutpoint (outpoint: Outpoint): ExitOperation {
    this.outpoint = outpoint;
    return this;
  }

  /**
   * Sets the exit bond.
   *
   * @param committedFee The amount of the fee.
   */
  withCommittedFee (committedFee: BN): ExitOperation {
    this.committedFee = committedFee;
    return this;
  }

  /**
   * Performs the exit. Will throw an error if `outpoint` or
   * `committedFee` aren't set.
   *
   * Performing an exit entails the following steps:
   *
   * 1. Retrieving all transactions from the block referenced by `outpoint`.
   * 2. Generating a Merkle proof-of-inclusion for the exiting transaction.
   * 3. Starting the exit on the root chain.
   */
  async exit () {
    assert(this.outpoint, 'an outpoint to exit must be provided');
    assert(this.committedFee, 'a fee must be provided');
    if (this.outpoint!.amount.lt(this.committedFee!)) {
      throw new Error('outpoint cannot be smaller than the committed fee');
    }

    const block = await this.client.getBlock(this.outpoint!.blockNum);
    const merkle = new MerkleTree();
    for (const tx of block.transactions) {
      merkle.addItem(sha256(tx.transaction.toRLP()));
    }
    const {proof} = merkle.generateProofAndRoot(this.outpoint!.txIdx);
    const confirmSigs: [Buffer, Buffer] = [
      this.outpoint!.confirmSig,
      Buffer.from(''),
    ];
    await this.contract.startExit(this.outpoint!, proof, confirmSigs, this.committedFee!, this.from);
  }
}
