import * as protoLoader from '@grpc/proto-loader';
import * as grpc from 'grpc';
import * as path from 'path';
import {pify} from './pify';
import {GetBalanceResponse, GetBlockResponse, GetOutputsResponse} from './PlasmaRPC';
import {toBig} from './numbers';
import {parseHex} from './parseHex';
import BN = require('bn.js');

export type ClientCB<T> = (err: any, res: T) => void;

export interface IClient {
  getBalance (args: { address: Buffer }, cb: ClientCB<GetBalanceResponse>): void

  getBlock (args: { number: number }, cb: ClientCB<GetBlockResponse>): void

  getOutputs (args: { address: Buffer, spendable: boolean }, cb: ClientCB<GetOutputsResponse>): void
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

  public async getBlock (number: number): Promise<GetBlockResponse> {
    return pify<GetBlockResponse>((cb) => this.client.getBlock({number}, cb));
  }

  public async getUTXOs (address: string): Promise<GetOutputsResponse> {
    return pify<GetOutputsResponse>((cb) => this.client.getOutputs({address: parseHex(address), spendable: true}, cb));
  }

  static getShared (): PlasmaClient {
    if (cachedClient) {
      return cachedClient;
    }

    cachedClient = new PlasmaClient(path.resolve(__dirname, '..', '..', '..', 'rpc', 'proto', 'root.proto'), 'localhost:6545');
    return cachedClient;
  }
}