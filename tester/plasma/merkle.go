package plasma

import (
	"math/big"

	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/util"
)

// TODO: move to merkle in main code

func CreateMerkleProof(merkle util.MerkleTree, index *big.Int) []byte {
	proofs := FindProofs(&merkle.Root, [][]byte{}, 1)

	if index.Int64() >= int64(len(proofs)) {
		panic("Transaction index must be within set of proofs")
	}

	return proofs[index.Int64()]
}

// TODO: we could optimize this with an index.
func FindProofs(node *util.MerkleNode, curr [][]byte, depth int) [][]byte {
	if node.Left == nil && node.Right == nil {
		if depth == 16 {
			// Reverse it.
			var copyCurr []byte

			for i := len(curr) - 1; i >= 0; i-- {
				copyCurr = append(copyCurr, curr[i]...)
			}

			return [][]byte{copyCurr}
		}

		return [][]byte{}
	}

	var left [][]byte
	var right [][]byte

	if node.Left != nil {
		left = FindProofs(node.Left, append(curr, node.Right.Hash), depth+1)
	}

	if node.Right != nil {
		right = FindProofs(node.Right, append(curr, node.Left.Hash), depth+1)
	}

	return append(left, right...)
}

func CreateMerkleTree(accepted []chain.Transaction) util.MerkleTree {
	hashables := make([]util.RLPHashable, len(accepted))

	for i := range accepted {
		txPtr := &accepted[i]
		hashables[i] = util.RLPHashable(txPtr)
	}

	merkle := util.TreeFromRLPItems(hashables)
	return merkle
}
