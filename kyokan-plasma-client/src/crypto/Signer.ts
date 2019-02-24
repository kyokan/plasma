import {ethSign} from './sign';
import {parseHex, toHex} from '../util/hex';
import Web3 = require('web3');

export interface Signer {
  ethSign (hash: Buffer): Promise<Buffer>
}

export class LocalSigner implements Signer {
  private readonly privateKey: Buffer;

  constructor (privateKey: Buffer) {
    this.privateKey = privateKey;
  }

  async ethSign (hash: Buffer): Promise<Buffer> {
    return ethSign(hash, this.privateKey);
  }
}

export class PersonalSigner implements Signer {
  private readonly web3: Web3;

  private readonly from: string;

  constructor (web3: Web3, from: string) {
    this.web3 = web3;
    this.from = from;
  }

  async ethSign (hash: Buffer): Promise<Buffer> {
    const message = toHex(hash);
    return new Promise<Buffer>((resolve, reject) => (this.web3.currentProvider as any).sendAsync({
      method: 'personal_sign',
      params: [message, this.from],
      from: this.from,
    }, (err: any, result: { result: string }) => {
      if (err) {
        return reject(err);
      }

      resolve(parseHex(result.result));
    }));
  }
}