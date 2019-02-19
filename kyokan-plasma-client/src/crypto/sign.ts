import * as ejs from 'ethereumjs-util';
import {gethKeccak256} from './hash';

export function sign (hash: Buffer, privateKey: Buffer): Buffer {
  if (hash.length !== 32) {
    throw new Error('Hash must be of length 32.');
  }

  const sig = ejs.ecsign(hash, privateKey);
  return Buffer.concat([sig.r, sig.s, Buffer.from([sig.v])]);
}

export function ethSign (hash: Buffer, privateKey: Buffer): Buffer {
  return sign(gethKeccak256(hash), privateKey);
}