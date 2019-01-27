import Web3 from 'web3';
import {Config} from '../Config';
import {SharedWeb3} from './SharedWeb3';
import BN = require('bn.js');
import abi from '../abi/PlasmaMVP.abi.json';
import Contract from 'web3/eth/contract';
import {TransactionReceipt} from 'web3/types';
import {Tx} from 'web3/eth/types';
import PromiEvent from 'web3/promiEvent';

let cachedContract: PlasmaContract;

export default class PlasmaContract {
  private web3: Web3;

  private contract: Contract;

  constructor (web3: Web3) {
    this.web3 = web3;
    this.contract = new this.web3.eth.Contract(abi, Config.PLASMA_CONTRACT_ADDRESS);
  }

  public deposit (value: BN, from: string): Promise<TransactionReceipt> {
    return this.awaitReceipt(() => this.contract.methods.deposit(from).send(this.decorateCall({
      to: this.contract.options.address,
      value: value.toString(10),
      from,
    })));
  }

  private decorateCall(args: Tx): Tx {
    return {
      ...args,
      gas: '1000000'
    };
  }

  private awaitReceipt(cb: () => PromiEvent<any>): Promise<TransactionReceipt> {
    return new Promise((resolve, reject) => {
      cb().on('receipt', resolve)
        .on('error', reject);
    })
  }

  static getShared (): PlasmaContract {
    if (cachedContract) {
      return cachedContract;
    }

    cachedContract = new PlasmaContract(SharedWeb3.getShared());
    return cachedContract;
  }
}