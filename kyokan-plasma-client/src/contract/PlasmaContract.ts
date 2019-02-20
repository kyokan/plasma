import Web3 from 'web3';
import abi from '../abi/PlasmaMVP.abi.json';
import {EventLog, TransactionReceipt} from 'web3/types';
import {Tx} from 'web3/eth/types';
import PromiEvent from 'web3/promiEvent';
import {PlasmaMVP} from '../abi/PlasmaMVP';
import Outpoint from '../domain/Outpoint';
import {toHex} from '../util/hex';
import {OnChainDeposit} from '../domain/OnChainDeposit';
import {toBig} from '../util/numbers';
import BN = require('bn.js');

interface DepositEvent {
  nonce: string
  owner: string;
  amount: string;
  createdAt: string;
  ethBlockNum: string;
}

export default class PlasmaContract {
  private web3: Web3;

  private contract: PlasmaMVP;

  constructor (web3: Web3, address: string) {
    this.web3 = web3;
    this.contract = new this.web3.eth.Contract(abi, address) as PlasmaMVP;
  }

  public async deposit (value: BN, from: string): Promise<OnChainDeposit> {
    const receipt = await this.awaitReceipt(() => this.contract.methods.deposit(from).send(this.decorateCall({
      to: this.contract.options.address,
      value: value.toString(10),
      from,
    })));

    const ev = receipt.events!.Deposit.returnValues as DepositEvent;

    return {
      nonce: toBig(ev.nonce),
      owner: ev.owner,
      amount: toBig(ev.amount),
      createdAt: ev.createdAt,
      ethBlockNum: ev.ethBlockNum,
    };
  }

  public async depositNonce (): Promise<BN> {
    const nonce = await this.contract.methods.depositNonce().call();
    return new BN(nonce);
  }

  public async depositFor (nonce: BN): Promise<OnChainDeposit> {
    const deposit = await this.contract.methods.deposits(nonce.toString(10)).call();
    return {
      nonce,
      owner: deposit.owner,
      amount: new BN(deposit.amount),
      createdAt: deposit.createdAt,
      ethBlockNum: deposit.ethBlockNum,
    };
  }

  public startExit (outpoint: Outpoint, proof: Buffer, confirmSignatures: [Buffer, Buffer], fee: BN, from: string): Promise<TransactionReceipt> {
    if (!outpoint.transaction) {
      throw new Error('exiting outpoint must have associated transaction');
    }

    return this.awaitReceipt(() => this.contract.methods.startTransactionExit(
      [
        outpoint.blockNum,
        outpoint.txIdx,
        outpoint.outIdx,
      ],
      toHex(outpoint.transaction!.transaction.toRLP()),
      toHex(proof),
      toHex(Buffer.concat(confirmSignatures)),
      fee.toString(10),
    ).send(this.decorateCall({
      to: this.contract.options.address,
      value: fee.toString(10),
      from,
    })));
  }

  public startDepositExit (nonce: BN, fee: BN, from: string): Promise<TransactionReceipt> {
    return this.awaitReceipt(() => this.contract.methods.startDepositExit(nonce.toString(10), fee.toString(10)).send(this.decorateCall({
      to: this.contract.options.address,
      value: fee.toString(10),
      from,
    })));
  }

  public challengedExits (): Promise<EventLog[]> {
    return this.contract.getPastEvents('ChallengedExit', {
      fromBlock: 0,
    });
  }

  private decorateCall (args: Tx): Tx {
    return {
      ...args,
      gas: '1000000',
    };
  }

  private awaitReceipt (cb: () => PromiEvent<any>): Promise<TransactionReceipt> {
    return new Promise((resolve, reject) => {
      cb().on('receipt', resolve)
        .on('error', reject);
    });
  }
}
