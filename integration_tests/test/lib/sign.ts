import * as ejs from 'ethereumjs-util';
import {keccak256} from './hash';

export function sign (hash: Buffer, privateKey: Buffer): Buffer {
  const sig = ejs.ecsign(hash, privateKey);
  return Buffer.concat([sig.r, sig.s, Buffer.from([sig.v])]);
}

export function ethSign(hash: Buffer, privateKey: Buffer): Buffer {
  const sig = ejs.ecsign(keccak256(Buffer.concat([
    Buffer.from('\x19', 'utf-8'),
    Buffer.from('Ethereum Signed Message:\n32', 'utf-8'),
    hash
  ])), privateKey);
  return Buffer.concat([sig.r, sig.s, Buffer.from([sig.v])]);
}