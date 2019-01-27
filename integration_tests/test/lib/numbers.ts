import BN = require('bn.js');
import {BNWire} from './PlasmaRPC';
import * as ejs from 'ethereumjs-util';

export function toBig(num: string|number|BN): BN {
  if (num instanceof BN) {
    return num;
  }

  if (typeof num === 'string' && num.indexOf('0x') === 0) {
    return new BN(num.slice(2), 16);
  }

  return new BN(num);
}

export function toBuffer(num: string|number|BN|BNWire, bufLen: number = 32): Buffer {
  if (num instanceof BN) {
    return num.toBuffer('be', bufLen);
  }

  if (typeof (num as any).hex === 'string') {
    num = (num as any).hex
  }

  const buf = ejs.toBuffer(num);
  const diff = length - buf.length;
  return Buffer.concat([ Buffer.from(new Uint8Array(diff)), buf]);
}
