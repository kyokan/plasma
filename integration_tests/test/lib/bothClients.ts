import IRootClient from 'kyokan-plasma-client/lib/rpc/IRootClient';
import {SharedGRPCRootClient, SharedRESTRootClient} from './SharedRootClients';

export default function bothClients(cb: (client: IRootClient) => void) {
  describe('With the GRPC client', () => cb(SharedGRPCRootClient.get()));
  describe('With the REST client', () => cb(SharedRESTRootClient.get()));
}