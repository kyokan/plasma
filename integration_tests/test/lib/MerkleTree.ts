import {tmSHA256} from './hash';

export default class MerkleTree {
  private leaves: Buffer[] = [];

  public addItem (leaf: Buffer) {
    this.leaves.push(leaf);
  }

  public generateProofAndRoot (index: number): { root: Buffer, proof: Buffer } {
    const [root, proof] = this.doProofAndRoot(this.leaves, index);
    return {
      root,
      proof,
    };
  }

  private doProofAndRoot (leaves: Buffer[], index: number): [Buffer, Buffer] {
    if (leaves.length === 0) {
      return [Buffer.from(''), Buffer.from('')];
    }

    if (leaves.length === 1) {
      return [leaves[0], Buffer.from('')];
    }

    let pivot = Math.floor((leaves.length + 1) / 2);
    let proof = Buffer.from('');
    let left, right;

    if (index < pivot) {
      left = this.doProofAndRoot(leaves.slice(0, pivot), index);
      right = this.doProofAndRoot(leaves.slice(pivot), -1);

      if (index >= 0) {
        proof = Buffer.concat([left[1], right[0]]);
      }
    } else {
      left = this.doProofAndRoot(leaves.slice(0, pivot), -1);
      right = this.doProofAndRoot(leaves.slice(pivot), index - pivot);

      if (index >= 0) {
        proof = Buffer.concat([right[1], left[0]]);
      }
    }

    return [tmSHA256([left[0], right[0]]), proof];
  }
}