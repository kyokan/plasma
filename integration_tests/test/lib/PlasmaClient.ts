import * as protoLoader from '@grpc/proto-loader';
import * as grpc from 'grpc';
import * as path from 'path';
import {pify} from './pify';
import {
  BlockWire,
  ConfirmedTransactionWire,
  GetBalanceResponse,
  GetConfirmationsResponse,
  GetOutputsResponse,
  SendResponse,
} from './PlasmaRPC';
import {toBig} from './numbers';
import {parseHex} from './parseHex';
import Outpoint from '../domain/Outpoint';
import Transaction from '../domain/Transaction';
import Block from '../domain/Block';
import {ethSign} from './sign';
import {keccak256} from './hash';
import BN = require('bn.js');

export type ClientCB<T> = (err: any, res: T) => void;

export interface IClient {
  getBalance (args: { address: Buffer }, cb: ClientCB<GetBalanceResponse>): void

  getBlock (args: { number: number }, cb: ClientCB<BlockWire>): void

  getOutputs (args: { address: Buffer, spendable: boolean }, cb: ClientCB<GetOutputsResponse>): void

  send (args: any, cb: ClientCB<SendResponse>): void

  confirm (args: any, cb: ClientCB<any>): void

  getConfirmations (args: any, cb: ClientCB<GetConfirmationsResponse>): void
}

let cachedClient: PlasmaClient;

export default class PlasmaClient {
  private client: IClient;

  constructor (protoFile: string, url: string) {
    const definition = protoLoader.loadSync(
      protoFile,
      {
        keepCase: true,
        longs: String,
        enums: String,
        defaults: true,
        oneofs: true,
      },
    );
    const pb = grpc.loadPackageDefinition(definition).pb;
    this.client = new (pb as any).Root(url, grpc.credentials.createInsecure());
  }

  public async getBalance (address: string): Promise<BN> {
    const res = await pify<GetBalanceResponse>((cb) => this.client.getBalance({address: parseHex(address)}, cb));
    return toBig(res.balance.hex);
  }

  public async getBlock (number: number): Promise<Block> {
    const blockWire = await pify<BlockWire>((cb) => this.client.getBlock({number}, cb));
    return Block.fromWire(blockWire);
  }

  public async getUTXOs (address: string): Promise<Outpoint[]> {
    const res = await pify<GetOutputsResponse>((cb) => this.client.getOutputs({
      address: parseHex(address),
      spendable: true,
    }, cb));
    return res.confirmedTransactions.map((r: ConfirmedTransactionWire) => Outpoint.fromWireTx(r, address));
  }

  public async send (tx: Transaction, confirmSigs: Buffer[]): Promise<SendResponse['inclusion']> {
    const res = await pify<SendResponse>((cb) => this.client.send({confirmed: tx.toRPC(confirmSigs)}, cb));
    res.inclusion.blockNumber = Number(res.inclusion.blockNumber);
    return res.inclusion;
  }

  public async confirm (blockNumber: number, txIdx: number, authSigs: [Buffer, Buffer]) {
    return pify((cb) => this.client.confirm({
      blockNumber,
      transactionIndex: txIdx,
      authSig0: authSigs[0],
      authSig1: authSigs[1],
    }, cb));
  }

  public async getConfirmations (privateKey: Buffer, blockNumber: number, txIdx: number, outIdx: number) {
    const now = Math.floor(Date.now() / 1000);
    const buf = Buffer.concat([
      Buffer.from(now.toString(), 'utf-8'),
      Buffer.from('kyo-plasma-mvp', 'utf-8'),
    ]);
    const hash = keccak256(buf);
    const sig = ethSign(hash, privateKey);
    return pify<GetConfirmationsResponse>((cb) => this.client.getConfirmations({
      blockNumber,
      transactionIndex: txIdx,
      outputIndex: outIdx,
      sig,
      nonce: now,
    }, cb));
  }

  static getShared (): PlasmaClient {
    if (cachedClient) {
      return cachedClient;
    }

    cachedClient = new PlasmaClient(path.resolve(__dirname, '..', '..', '..', 'rpc', 'proto', 'root.proto'), 'localhost:6545');
    return cachedClient;
  }
}