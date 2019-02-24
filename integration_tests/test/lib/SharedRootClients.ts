import IRootClient from 'kyokan-plasma-client/lib/rpc/IRootClient';
import {RESTRootClient} from 'kyokan-plasma-client/lib/rpc/RESTRootClient';
import GRPCRootClient from 'kyokan-plasma-client/lib/rpc/GRPCRootClient';

let cachedGRPC: IRootClient;
let cachedREST: IRootClient;

export class SharedGRPCRootClient {
  static get () {
    if (!cachedGRPC) {
      cachedGRPC = new GRPCRootClient('localhost:6545');
    }

    return cachedGRPC;
  }
}

export class SharedRESTRootClient {
  static get () {
    if (!cachedREST) {
      cachedREST = new RESTRootClient('http://localhost:6546');
    }

    return cachedREST;
  }
}