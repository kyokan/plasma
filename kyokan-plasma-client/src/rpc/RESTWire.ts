import Block from '../domain/Block';
import Input from '../domain/Input';
import Output from '../domain/Output';
import {parseHex, toHex} from '../util/hex';
import TransactionBody from '../domain/TransactionBody';
import Transaction from '../domain/Transaction';
import ConfirmedTransaction from '../domain/ConfirmedTransaction';
import Outpoint from '../domain/Outpoint';
import BN = require('bn.js');
import {addressesEqual} from '../util/addresses';

export interface BlockWire {
  block: {
    header: {
      merkleRoot: string
      rlpMerkleRoot: string
      prevHash: string
      number: number
    },
    hash: string
  },
  confirmedTransactions: ConfirmedTransactionWire[]
}

export function blockFromWire (blockWire: BlockWire): Block {
  return new Block(
    {
      merkleRoot: parseHex(blockWire.block.header.merkleRoot),
      number: Number(blockWire.block.header.number),
    },
    blockWire.confirmedTransactions.map((ct) => confirmedTransactionFromWire(ct)),
  );
}

export interface InputWire {
  depositNonce: string
  blockNumber: number
  transactionIndex: number
  outputIndex: number
}

export function inputFromWire (input: InputWire): Input {
  return new Input(
    input.blockNumber,
    input.transactionIndex,
    input.outputIndex,
    input.depositNonce !== '0' ? new BN(input.depositNonce) : undefined,
  );
}

export function inputToWire (input: Input): InputWire {
  return {
    blockNumber: input.blockNum,
    transactionIndex: input.txIdx,
    outputIndex: input.outIdx,
    depositNonce: input.depositNonce.toString(10),
  };
}

export interface OutputWire {
  owner: string
  amount: string
}

export function outputFromWire (output: OutputWire): Output {
  return new Output(
    output.owner,
    new BN(output.amount),
  );
}

export function outputToWire (output: Output): OutputWire {
  return {
    owner: output.owner,
    amount: output.amount.toString(10),
  };
}

export interface TransactionBodyWire {
  input0: InputWire
  input0ConfirmSig: string
  input1: InputWire
  input1ConfirmSig: string
  output0: OutputWire
  output1: OutputWire
  fee: string
  blockNumber: number
  transactionIndex: number
}

export function transactionBodyFromWire (tx: TransactionBodyWire): TransactionBody {
  return new TransactionBody(
    inputFromWire(tx.input0),
    inputFromWire(tx.input1),
    outputFromWire(tx.output0),
    outputFromWire(tx.output1),
    tx.blockNumber,
    tx.transactionIndex,
    parseHex(tx.input0ConfirmSig),
    parseHex(tx.input1ConfirmSig),
    new BN(tx.fee),
  );
}

export function transactionBodyToWire (tx: TransactionBody): TransactionBodyWire {
  return {
    blockNumber: tx.blockNum,
    transactionIndex: tx.txIdx,
    input0ConfirmSig: toHex(tx.input0ConfirmSig),
    input1ConfirmSig: toHex(tx.input1ConfirmSig),
    fee: tx.fee.toString(10),
    input0: inputToWire(tx.input0),
    input1: inputToWire(tx.input1),
    output0: outputToWire(tx.output0),
    output1: outputToWire(tx.output1),
  };
}

export interface TransactionWire {
  body: TransactionBodyWire
  sigs: string[]
}

export function transactionFromWire (tx: TransactionWire): Transaction {
  return new Transaction(
    transactionBodyFromWire(tx.body),
    parseHex(tx.sigs[0]),
    parseHex(tx.sigs[1]),
  );
}

export function transactionToWire (tx: Transaction): TransactionWire {
  return {
    body: transactionBodyToWire(tx.body),
    sigs: [
      toHex(tx.signature0!),
      toHex(tx.signature0!),
    ],
  };
}

export interface ConfirmedTransactionWire {
  confirmSigs: string[]
  transaction: TransactionWire
}

export function confirmedTransactionFromWire (wireTx: ConfirmedTransactionWire): ConfirmedTransaction {
  return new ConfirmedTransaction(
    transactionFromWire(wireTx.transaction),
    [parseHex(wireTx.confirmSigs[0]), parseHex(wireTx.confirmSigs[1])],
  );
}

export function outpointFromConfirmedTxWire (txWire: ConfirmedTransactionWire, owner: string): Outpoint {
  const tx = confirmedTransactionFromWire(txWire);
  const body = tx.transaction.body;
  const outIdx = addressesEqual(owner, body.output0.owner) ? 0 : 1;

  if (!tx.confirmSignatures) {
    throw new Error('cannot create outpoint from unconfirmed transaction');
  }

  return new Outpoint(
    body.blockNum,
    body.txIdx,
    outIdx,
    outIdx === 0 ? body.output0.amount : body.output1.amount,
    tx.confirmSignatures[outIdx],
    tx,
  );
}

export interface SendResponseWire {
  transaction: TransactionWire,
  inclusion: {
    merkleRoot: string,
    blockNumber: number,
    transactionIndex: number,
  }
}