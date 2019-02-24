import IRootClient from './IRootClient';
import ConfirmedTransaction from '../domain/ConfirmedTransaction';
import Block from '../domain/Block';
import Outpoint from '../domain/Outpoint';
import {SendResponse} from '../domain/SendResponse';
import Transaction from '../domain/Transaction';
import {
  blockFromWire,
  BlockWire,
  ConfirmedTransactionWire,
  outpointFromConfirmedTxWire,
  SendResponseWire,
  transactionFromWire,
  transactionToWire,
} from './RESTWire';
import {parseHex, toHex} from '../util/hex';
import BN = require('bn.js');

export class RESTRootClient implements IRootClient {
  private readonly rootURL: string;

  constructor (rootURL: string) {
    this.rootURL = rootURL;
  }

  public async confirm (confirmed: ConfirmedTransaction): Promise<void> {
    const req = {
      blockNumber: confirmed.transaction.body.blockNum,
      transactionIndex: confirmed.transaction.body.txIdx,
      confirmSig0: toHex(confirmed.confirmSignatures![0]),
      confirmSig1: toHex(confirmed.confirmSignatures![1]),
    };

    await this.doRequest<any>('confirm', 'POST', req);
  }

  public async getBalance (address: string): Promise<BN> {
    const res = await this.doRequest<{ balance: string }>(`balances/${address}`, 'GET');
    return new BN(res.balance);
  }

  public async getBlock (number: number): Promise<Block> {
    const res = await this.doRequest<BlockWire>(`blocks/${number}`, 'GET');
    return blockFromWire(res);
  }

  public async getUTXOs (address: string): Promise<Outpoint[]> {
    const res = await this.doRequest<ConfirmedTransactionWire[]>(`utxos/${address}`, 'GET');
    return res.map((r) => outpointFromConfirmedTxWire(r, address));
  }

  public async send (tx: Transaction): Promise<SendResponse> {
    const wire = transactionToWire(tx);
    const res = await this.doRequest<SendResponseWire>('send', 'POST', wire);

    return {
      transaction: transactionFromWire(res.transaction),
      inclusion: {
        merkleRoot: parseHex(res.inclusion.merkleRoot),
        blockNumber: res.inclusion.blockNumber,
        transactionIndex: res.inclusion.transactionIndex,
      },
    };
  }

  private async doRequest<T> (path: string, method: 'GET' | 'POST', body?: any): Promise<T> {
    const opts = {
      method,
      mode: 'cors',
    } as any;
    if (body) {
      opts.headers = {
        'Content-Type': 'application/json',
      };
      opts.body = JSON.stringify(body);
    }

    const res = await fetch(`${this.rootURL}/${path}`, opts);
    if (res.status !== 200) {
      throw new Error(`received bad status code: ${res.status}`);
    }
    const json = await res.json();
    return json as T;
  }
}