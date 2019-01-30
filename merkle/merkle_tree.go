package merkle

import (
	"encoding/binary"
	"github.com/kyokan/plasma/util"
	"math/big"
)

type DualHashable interface {
	util.Hashable
	util.RLPHashable
	SetIndex(uint32)
	GetFee() *big.Int
}

func getMerkleRoot(hasher util.Hasher, leaves []DualHashable) ([]byte, []byte) {
	switch len(leaves) {
	case 0:
		return nil, nil
	case 1:
		return leaves[0].RLPHash(hasher), leaves[0].Hash(hasher)
	default:
		mid := (1 + len(leaves)) / 2
		lhsRLP, lhs := getMerkleRoot(hasher, leaves[:mid])
		rhsRLP, rhs := getMerkleRoot(hasher, leaves[mid:])
		return hasher(encodeLength(lhsRLP, rhsRLP)), hasher(encodeLength(lhs, rhs))
	}
}

func encodeLength(lhs, rhs []byte) []byte {
	var lbuf, rbuf [10]byte
	lsize := binary.PutUvarint(lbuf[:], uint64(len(lhs)))
	rsize := binary.PutUvarint(rbuf[:], uint64(len(rhs)))
	left := append(lbuf[0:lsize], lhs...)
	right := append(rbuf[0:rsize], rhs...)
	return append(left, right...)
}

func GetMerkleRoot(leaves []DualHashable) ([]byte, []byte) {
	return getMerkleRoot(util.Sha256, leaves)
}
