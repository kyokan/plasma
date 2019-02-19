import Web3 = require('web3');
import GRPCRootClient from './rpc/GRPCRootClient';
import PlasmaContract from './contract/PlasmaContract';
import * as grpc from 'grpc';
import {ChannelCredentials} from 'grpc';
import {parseHex, toHex} from './util/hex';
import {NumberLike} from './domain/Numbers';
import ConfirmedTransaction from './domain/ConfirmedTransaction';
import SendOperation from './operations/SendOperation';
import * as ejs from 'ethereumjs-util';
import {toBig} from './util/numbers';
import {OnChainDeposit} from './domain/OnChainDeposit';
import IRootClient from './rpc/IRootClient';

/**
 * Options used to construct a new Plasma client instance.
 */
export interface PlasmaOptions {
  /**
   * An instance of Web3. To transact, your Web3 instance
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
   * NOTE: Currently unsupported.
   *
   * Whether to use the `grpc-web` Client rather than the default
   * `grpc` client. Set this to `true` if you're connecting to Plasma
   * from a web browser.
   */
  useGRPCWeb: boolean

  /**
   * URL to the root node. The URL should not include a protocol.
   *
   * @example localhost:6545
   */
  rootURL: string

  /**
   * gRPC credentials to use to connect to the root node. Will default
   * to HTTP transport.
   */
  rootCredentials?: ChannelCredentials

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

  private readonly privateKey: Buffer | null = null;

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

    if (opts.useGRPCWeb) {
      throw new Error('gRPC Web support coming soon.');
    } else {
      this._rootClient = new GRPCRootClient(opts.rootURL, opts.rootCredentials || grpc.credentials.createInsecure());
    }

    this._contract = new PlasmaContract(this.web3, opts.contractAddress);

    if (opts.privateKey) {
      if (typeof opts.privateKey === 'string') {
        this.privateKey = parseHex(opts.privateKey);
      } else {
        this.privateKey = opts.privateKey;
      }

      this.fromAddress = toHex(ejs.privateToAddress(this.privateKey) as Buffer);
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

    return sendOp.send(this.privateKey!);
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

  private ensureKey () {
    if (!this.privateKey || !this.fromAddress) {
      throw new Error('must set a private key or from address to perform this operation');
    }
  }
}