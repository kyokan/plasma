import Web3 = require('web3');
import PlasmaContract from './contract/PlasmaContract';
import {parseHex, toHex} from './util/hex';
import {NumberLike} from './domain/Numbers';
import ConfirmedTransaction from './domain/ConfirmedTransaction';
import SendOperation from './operations/SendOperation';
import * as ejs from 'ethereumjs-util';
import {toBig} from './util/numbers';
import {OnChainDeposit} from './domain/OnChainDeposit';
import IRootClient from './rpc/IRootClient';
import Outpoint from './domain/Outpoint';
import ExitOperation from './operations/ExitOperation';
import {LocalSigner, PersonalSigner, Signer} from './crypto/Signer';

/**
 * Options used to construct a new Plasma client instance.
 */
export interface PlasmaOptions {
  /**
   * An instance of Web3. To transact, your Webpa3 instance
   * must support signing using the same `privateKey` provided
   * to the Plasma client.
   */
  web3: Web3

  /**
   * The address of the Plasma smart contract on the root
   * chain.
   */
  contractAddress: string

  /**
   * The client to connect to the root node with. Two clients are included
   * in this package:
   *
   * - GRPCRootClient: A client that uses gRPC to connect to the root client. You
   * should use this if you're running a server-side application.
   * - RESTRootClient: A client that uses a REST api to connect to the root client.
   * You should use this if you're running a browser application.
   */
  rootClient: IRootClient

  /**
   * Address to sign with and deposit from.
   */
  fromAddress: string

  /**
   * Your private key, either hex-encoded or as a Buffer.
   */
  privateKey?: string | Buffer
}

/**
 * Entrypoint to the Plasma client. Provides convenience methods for common
 * functionality such as sending value.
 */
export default class Plasma {
  private readonly web3: Web3;

  private readonly _rootClient: IRootClient;

  private readonly _contract: PlasmaContract;

  private readonly signer: Signer;

  private readonly fromAddress: string | null = null;

  /**
   * Constructs a new Plasma client. All options are required, but
   * `privateKey` can be null. If you don't set a `privateKey`, methods
   * that require it such as `send` will throw an error.
   *
   * See [[PlasmaOptions]] for the full list of supported options.
   *
   * @param opts
   */
  constructor (opts: PlasmaOptions) {
    this.web3 = opts.web3;
    this._rootClient = opts.rootClient;

    this._contract = new PlasmaContract(this.web3, opts.contractAddress);

    if (opts.privateKey) {
      let priv: Buffer;
      if (typeof opts.privateKey === 'string') {
        priv = parseHex(opts.privateKey);
      } else {
        priv = opts.privateKey;
      }

      this.signer = new LocalSigner(priv);
      this.fromAddress = toHex(ejs.privateToAddress(priv) as Buffer);
    } else {
      this.fromAddress = opts.fromAddress;
      this.signer = new PersonalSigner(this.web3, this.fromAddress);
    }
  }

  /**
   * Accessor method for the root node's gRPC client. Use this to
   * query balances, list UTXOs, and make other requests to the
   * root node directly.
   */
  public rootNode (): IRootClient {
    return this._rootClient;
  }

  /**
   * Accessor method for the Plasma smart contract. Use this to
   * make smart contract calls directly.
   */
  public contract (): PlasmaContract {
    return this._contract;
  }

  /**
   * Sends funds.
   *
   * @param to Address to send funds to.
   * @param value How much money to send.
   * @param fee A fee amount.
   * @param depositNonce The deposit nonce you'd like to spend.
   */
  public async send (to: string, value: NumberLike, fee: NumberLike, depositNonce?: NumberLike): Promise<ConfirmedTransaction> {
    this.ensureKey();

    const bigVal = toBig(value);
    const bigFee = toBig(fee);

    const sendOp = new SendOperation(this.rootNode(), this.contract(), this.fromAddress!)
      .toAddress(to)
      .forValue(bigVal)
      .withFee(bigFee);

    if (depositNonce) {
      sendOp.withDepositNonce(toBig(depositNonce));
    }

    return sendOp.send(this.signer!);
  }

  /**
   * Deposits funds into the Plasma smart contract. Pass the `depositNonce`
   * included as part of the returned the [[OnChainDeposit]] to `send` to
   * spend the deposited funds.
   *
   * @param value Amount you'd like to deposit.
   */
  public async deposit (value: NumberLike): Promise<OnChainDeposit> {
    this.ensureKey();
    return this.contract().deposit(toBig(value), this.fromAddress!);
  }

  /**
   * Starts exiting the provided value or outpoint. If a value is provided,
   * the outpoint whose amount is closest but higher than the given value is
   * chosen. For example, if you have two UTXOs with amount 1.5ETH and 2ETH
   * and you call startExit with 1ETH, the 1.5ETH outpoint will be exited.
   *
   * The bond parameter is used to to protect against double-spends. Your bond
   * must be higher than the value of your outpoint. The value of the bond will
   * be returned when the challenge period expires.
   *
   * WARNING: You will forfeit your bond if you attempt to exit a spent outpoint.
   *
   * @param valueOrOutpoint The or outpoint to exit.
   * @param bond The bond to post guaranteeing this exit.
   */
  public async startExit (valueOrOutpoint: NumberLike | Outpoint, bond: NumberLike): Promise<void> {
    this.ensureKey();
    let outpoint: Outpoint | null = null;

    if (valueOrOutpoint instanceof Outpoint) {
      outpoint = valueOrOutpoint;
    } else {
      const value = toBig(valueOrOutpoint);
      const utxos = await this._rootClient.getUTXOs(this.fromAddress!);
      // outpoints are in descending-value order
      for (let i = utxos.length - 1; i >= 0; i--) {
        const utxo = utxos[i];
        if (utxo.amount.gte(value)) {
          outpoint = utxo;
          break;
        }
      }
    }

    if (!outpoint) {
      throw new Error('no suitable UTXOs');
    }

    const exitOp = new ExitOperation(this._contract, this._rootClient, this.fromAddress!)
      .withOutpoint(outpoint)
      .withCommittedFee(toBig(bond));

    await exitOp.exit();
  }

  private ensureKey () {
    if (!this.signer || !this.fromAddress) {
      throw new Error('must set a signer to perform this operation');
    }
  }
}