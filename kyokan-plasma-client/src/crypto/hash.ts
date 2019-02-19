import Web3 = require('web3');
import {parseHex} from '../util/hex';
import * as ejs from 'ethereumjs-util';

export function keccak256 (buf: Buffer): Buffer {
  return parseHex(Web3.utils.sha3(`0x${buf.toString('hex')}`));
}

export function sha256 (buf: Buffer): Buffer {
  return ejs.sha256(buf) as Buffer;
}

export function gethKeccak256 (hash: Buffer): Buffer {
  if (hash.length !== 32) {
    throw new Error('Hash must be of length 32.');
  }

  return keccak256(Buffer.concat([
    Buffer.from('\x19Ethereum Signed Message:\n32', 'utf-8'),
    hash,
  ]));
}

export function tmSHA256 (bufs: Buffer[]): Buffer {
  return sha256(Buffer.concat(bufs));
}