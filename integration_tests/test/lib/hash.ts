import Web3 = require('web3');
import {parseHex} from './parseHex';
import * as ejs from 'ethereumjs-util';

export function keccak256 (buf: Buffer): Buffer {
  return parseHex(Web3.utils.sha3(`0x${buf.toString('hex')}`));
}

export function sha256(buf: Buffer): Buffer {
  return ejs.sha256(buf) as Buffer;
}

export function tmSHA256(bufs: Buffer[]): Buffer {
  const items = [];
  for (const buf of bufs) {
    items.push(Buffer.from([buf.length]));
    items.push(buf);
  }

  return sha256(Buffer.concat(bufs));
}