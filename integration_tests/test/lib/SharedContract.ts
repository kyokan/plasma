import PlasmaContract from 'kyokan-plasma-client/lib/contract/PlasmaContract';
import {SharedWeb3} from './SharedWeb3';
import {Config} from '../Config';

let cachedContract: PlasmaContract;

export default class SharedContract {
  static get () {
    if (!cachedContract) {
      cachedContract = new PlasmaContract(SharedWeb3.get(), Config.PLASMA_CONTRACT_ADDRESS);
    }

    return cachedContract;
  }
}