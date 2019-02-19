import BN = require('bn.js');
import * as ejs from 'ethereumjs-util';
import {NumberLike} from '../domain/Numbers';

export function toBig (num: NumberLike): BN {
  if (num instanceof BN) {
    return num;
  }

  if (typeof num === 'string' && num.indexOf('0x') === 0) {
    return new BN(num.slice(2), 16);
  }

  return new BN(num);
}

export function toBuffer (num: NumberLike, bufLen: number = 32): Buffer {
  if (num instanceof BN) {
    return num.toBuffer('be', bufLen);
  }

  if (typeof (num as any).hex !== 'undefined') {
    num = (num as any).hex;
  }

  const buf = ejs.toBuffer(num);
  const diff = bufLen - buf.length;
  return Buffer.concat([Buffer.from(new Uint8Array(diff)), buf]);
}
