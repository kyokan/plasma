import * as protoLoader from '@grpc/proto-loader';
import * as grpc from 'grpc';
import {ChannelCredentials} from 'grpc';
import * as path from 'path';
import {pify} from '../util/pify';
import {toBig} from '../util/numbers';
import {parseHex} from '../util/hex';
import Outpoint from '../domain/Outpoint';
import Transaction from '../domain/Transaction';
import Block from '../domain/Block';
import ConfirmedTransaction from '../domain/ConfirmedTransaction';
import {
  blockFromWire,
  BlockWire,
  ConfirmedTransactionWire,
  ConfirmRequest,
  GetBalanceResponse,
  GetConfirmationsResponse,
  GetOutputsResponse,
  outpointFromConfirmedTxWire,
  sendResponseFromWire,
  SendResponseWire,
  transactionToWire,
} from './GRPCWire';
import {SendResponse} from '../domain/SendResponse';
import IRootClient from './IRootClient';
import BN = require('bn.js');

type ClientCB<T> = (err: any, res: T) => void;

interface InternalGRPCClient {
  getBalance (args: { address: Buffer }, cb: ClientCB<GetBalanceResponse>): void

  getBlock (args: { number: number }, cb: ClientCB<BlockWire>): void

  getOutputs (args: { address: Buffer, spendable: boolean }, cb: ClientCB<GetOutputsResponse>): void

  send (args: any, cb: ClientCB<SendResponseWire>): void

  confirm (args: ConfirmRequest, cb: ClientCB<any>): void

  getConfirmations (args: any, cb: ClientCB<GetConfirmationsResponse>): void
}

/**
 * gRPC client for the root node. Uses raw gRPC (i.e., not gRPC web) and
 * is therefore unsuitable for browser use.
 *
 * See [[IRootClient]] for descriptions of each method.
 */
export default class GRPCRootClient implements IRootClient {
  private client: InternalGRPCClient;

  /**
   * Constructs a new GRPCRootClient.
   *
   * @param url URL to the root node (without a protocol).
   * @param creds gRPC credentials to use while communicating with the root node.
   */
  constructor (url: string, creds?: ChannelCredentials) {
    const definition = protoLoader.loadSync(
      path.join(__dirname, 'pb', 'root.proto'),
      {
        keepCase: true,
        longs: String,
        enums: String,
        defaults: true,
        oneofs: true,
      },
    );
    const pb = grpc.loadPackageDefinition(definition).pb;
    this.client = new (pb as any).Root(url, creds || grpc.credentials.createInsecure());
  }

  public async getBalance (address: string): Promise<BN> {
    const res = await pify<GetBalanceResponse>((cb) => this.client.getBalance({address: parseHex(address)}, cb));
    return toBig(res.balance.hex);
  }

  public async getBlock (number: number): Promise<Block> {
    const blockWire = await pify<BlockWire>((cb) => this.client.getBlock({number}, cb));
    return blockFromWire(blockWire);
  }

  public async getUTXOs (address: string): Promise<Outpoint[]> {
    const res = await pify<GetOutputsResponse>((cb) => this.client.getOutputs({
      address: parseHex(address),
      spendable: true,
    }, cb));
    return res.confirmedTransactions.map((r: ConfirmedTransactionWire) => outpointFromConfirmedTxWire(r, address));
  }

  public async send (tx: Transaction): Promise<SendResponse> {
    const res = await pify<SendResponseWire>((cb) => this.client.send({transaction: transactionToWire(tx)}, cb));
    res.inclusion.blockNumber = Number(res.inclusion.blockNumber);
    return sendResponseFromWire(res);
  }

  public async confirm (confirmedTx: ConfirmedTransaction): Promise<void> {
    if (!confirmedTx.confirmSignatures) {
      throw new Error('cannot confirm a transaction without confirm sigs');
    }

    return pify((cb) => this.client.confirm({
      blockNumber: confirmedTx.transaction.body.blockNum,
      transactionIndex: confirmedTx.transaction.body.txIdx,
      confirmSig0: confirmedTx.confirmSignatures![0],
      confirmSig1: confirmedTx.confirmSignatures![1],
    }, cb));
  }
}