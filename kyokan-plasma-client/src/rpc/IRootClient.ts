import BN = require('bn.js');
import Block from '../domain/Block';
import Outpoint from '../domain/Outpoint';
import Transaction from '../domain/Transaction';
import {SendResponse} from '../domain/SendResponse';
import ConfirmedTransaction from '../domain/ConfirmedTransaction';

export default interface IRootClient {
  getBalance(address: string): Promise<BN>

  getBlock(number: number): Promise<Block>

  getUTXOs(address: string): Promise<Outpoint[]>

  send (tx: Transaction): Promise<SendResponse>

  confirm (confirmed: ConfirmedTransaction): Promise<void>
}