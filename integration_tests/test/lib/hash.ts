import Web3 = require('web3');
import {parseHex} from './parseHex';

export function keccak256 (buf: Buffer): Buffer {
  return parseHex(Web3.utils.sha3(`0x${buf.toString('hex')}`));
}