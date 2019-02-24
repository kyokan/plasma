import {LocalSigner, Signer} from 'kyokan-plasma-client/lib/crypto/Signer';

export function signerFromStr (key: Buffer): Signer {
  return new LocalSigner(key);
}