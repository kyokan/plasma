import Outpoint from '../domain/Outpoint';
import BN = require('bn.js');

export function selectUTXOs (utxos: Outpoint[], total: BN): Outpoint[] {
  const sorted = utxos.sort((a, b) => a.amount.cmp(b.amount));
  const first = sorted[0];

  if (first.amount.cmp(total) === 1) {
    return [first];
  }

  for (let i = sorted.length - 1; i > 0; i--) {
    const out = sorted[i];
    const sum = first.amount.add(out.amount);
    if (sum.cmp(total) >= 0) {
      return [first, out];
    }
  }

  throw new Error('no suitable UTXOs found');
}