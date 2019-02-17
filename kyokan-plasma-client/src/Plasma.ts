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

export interface PlasmaOptions {
  web3: Web3
  contractAddress: string
  useGRPCWeb: boolean
  rootURL: string
  rootCredentials: ChannelCredentials
  privateKey: string | Buffer | null
}

export default class Plasma {
  private readonly web3: Web3;

  private readonly _rootClient: IRootClient;

  private readonly _contract: PlasmaContract;

  private readonly privateKey: Buffer | null = null;

  private readonly fromAddress: string | null = null;

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

  public rootNode (): IRootClient {
    return this._rootClient;
  }

  public contract (): PlasmaContract {
    return this._contract;
  }

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