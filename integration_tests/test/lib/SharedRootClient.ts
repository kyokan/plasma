import GRPCRootClient from 'kyokan-plasma-client/lib/rpc/GRPCRootClient';
import IRootClient from 'kyokan-plasma-client/lib/rpc/IRootClient';

let cached: IRootClient;

export default class SharedRootClient {
  static get () {
    if (!cached) {
      cached = new GRPCRootClient('localhost:6545');
    }

    return cached;
  }
}