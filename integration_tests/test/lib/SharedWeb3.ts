import Web3 from 'web3';

let cachedWeb3: Web3;

export class SharedWeb3 {
  static getShared(): Web3 {
    if (cachedWeb3) {
      return cachedWeb3;
    }

    cachedWeb3 = new Web3('http://localhost:8545');
    return cachedWeb3;
  }
}