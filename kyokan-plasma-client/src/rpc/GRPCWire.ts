import BN = require('bn.js');
import {toBig} from '../util/numbers';
import Block from '../domain/Block';
import ConfirmedTransaction from '../domain/ConfirmedTransaction';
import Input from '../domain/Input';
import Output from '../domain/Output';
import {parseHex, toHex} from '../util/hex';
import TransactionBody from '../domain/TransactionBody';
import Transaction from '../domain/Transaction';
import {SendResponse} from '../domain/SendResponse';
import Outpoint from '../domain/Outpoint';

export interface BNWire {
  hex: string
}

export function toBNWire (num: BN | number): BNWire {
  if (!(num instanceof BN)) {
    num = new BN(num);
  }

  return {
    hex: num.toString('hex'),
  };
}

export function fromBNWire (num: BNWire): BN {
  return toBig(num.hex);
}

export interface GetBalanceResponse {
  balance: BNWire
}

export interface BlockWire {
  block: {
    header: {
      merkleRoot: Buffer
      rlpMerkleRoot: Buffer
      prevHash: Buffer
      number: string
    },
    hash: Buffer
  },
  confirmedTransactions: ConfirmedTransactionWire[]
}

export function blockFromWire (blockWire: BlockWire): Block {
  return new Block(
    {
      merkleRoot: blockWire.block.header.merkleRoot,
      number: Number(blockWire.block.header.number),
    },
    blockWire.confirmedTransactions.map((ct) => confirmedTransactionFromWire(ct)),
  );
}

export interface InputWire {
  depositNonce: BNWire
  blockNum: string
  txIdx: number
  outIdx: number
}

export function inputFromWire (input: InputWire): Input {
  return new Input(
    Number(input.blockNum),
    input.txIdx,
    input.outIdx,
    fromBNWire(input.depositNonce),
  );
}

export function inputToWire (input: Input): InputWire {
  return {
    blockNum: input.blockNum.toString(),
    txIdx: input.txIdx,
    outIdx: input.outIdx,
    depositNonce: toBNWire(input.depositNonce),
  };
}

export interface OutputWire {
  owner: Buffer
  amount: BNWire
}

export function outputFromWire (output: OutputWire): Output {
  return new Output(
    toHex(output.owner),
    fromBNWire(output.amount),
  );
}

export function outputToWire (output: Output): OutputWire {
  return {
    owner: parseHex(output.owner),
    amount: toBNWire(output.amount),
  };
}

export interface TransactionBodyWire {
  input0: InputWire
  input0ConfirmSig0: Buffer
  input0ConfirmSig1: Buffer
  input1: InputWire
  input1ConfirmSig0: Buffer
  input1ConfirmSig1: Buffer
  output0: OutputWire
  output1: OutputWire
  fee: BNWire
  blockNum: string
  txIdx: number
}

export function transactionBodyFromWire (tx: TransactionBodyWire): TransactionBody {
  return new TransactionBody(
    inputFromWire(tx.input0),
    inputFromWire(tx.input1),
    outputFromWire(tx.output0),
    outputFromWire(tx.output1),
    Number(tx.blockNum),
    tx.txIdx,
    [tx.input0ConfirmSig0, tx.input0ConfirmSig1],
    [tx.input1ConfirmSig0, tx.input1ConfirmSig1],
    fromBNWire(tx.fee),
  );
}

export function transactionBodyToWire (tx: TransactionBody): TransactionBodyWire {
  return {
    blockNum: tx.blockNum.toString(),
    txIdx: tx.txIdx,
    input0ConfirmSig0: tx.input0ConfirmSigs[0],
    input0ConfirmSig1: tx.input0ConfirmSigs[1],
    input1ConfirmSig0: tx.input1ConfirmSigs[0],
    input1ConfirmSig1: tx.input1ConfirmSigs[1],
    fee: toBNWire(tx.fee),
    input0: inputToWire(tx.input0),
    input1: inputToWire(tx.input1),
    output0: outputToWire(tx.output0),
    output1: outputToWire(tx.output1),
  };
}

export interface TransactionWire {
  body: TransactionBodyWire
  sig0: Buffer
  sig1: Buffer
}

export function transactionFromWire (tx: TransactionWire): Transaction {
  return new Transaction(
    transactionBodyFromWire(tx.body),
    tx.sig0,
    tx.sig1,
  );
}

export function transactionToWire (tx: Transaction): TransactionWire {
  return {
    body: transactionBodyToWire(tx.body),
    sig0: tx.signature0!,
    sig1: tx.signature1!,
  };
}

export interface ConfirmedTransactionWire {
  confirmSig0: Buffer
  confirmSig1: Buffer
  transaction: TransactionWire
}

export function confirmedTransactionFromWire (wireTx: ConfirmedTransactionWire): ConfirmedTransaction {
  return new ConfirmedTransaction(
    transactionFromWire(wireTx.transaction),
    [wireTx.confirmSig0, wireTx.confirmSig1],
  );
}

export function outpointFromConfirmedTxWire (txWire: ConfirmedTransactionWire, owner: string): Outpoint {
  const tx = confirmedTransactionFromWire(txWire);
  const body = tx.transaction.body;
  const outIdx = owner === body.output0.owner ? 0 : 1;

  if (!tx.confirmSignatures) {
    throw new Error('cannot create outpoint from unconfirmed transaction');
  }

  return new Outpoint(
    body.blockNum,
    body.txIdx,
    outIdx,
    outIdx === 0 ? body.output0.amount : body.output1.amount,
    tx.confirmSignatures,
    tx,
  );
}

export interface GetOutputsResponse {
  confirmedTransactions: ConfirmedTransactionWire[]
}

export interface SendResponseWire {
  transaction: TransactionWire
  inclusion: {
    merkleRoot: Buffer
    blockNumber: number
    transactionIndex: number
  }
}

export function sendResponseFromWire (res: SendResponseWire): SendResponse {
  return {
    transaction: transactionFromWire(res.transaction),
    inclusion: res.inclusion,
  };
}

export interface GetConfirmationsResponse {
  authSig0: Buffer
  authSig1: Buffer
}

export interface ConfirmRequest {
  blockNumber: number
  transactionIndex: number
  confirmSig0: Buffer
  confirmSig1: Buffer
}