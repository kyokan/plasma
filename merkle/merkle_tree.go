package merkle

import (
	"encoding/binary"
	"github.com/kyokan/plasma/util"
)

type DualHashable interface {
	util.RLPHashable
}

func getMerkleRoot(hasher util.Hasher, leaves []DualHashable) ([]byte) {
	switch len(leaves) {
	case 0:
		return nil
	case 1:
		return leaves[0].RLPHash(hasher)
	default:
		mid := (1 + len(leaves)) / 2
		lhsRLP := getMerkleRoot(hasher, leaves[:mid])
		rhsRLP := getMerkleRoot(hasher, leaves[mid:])
		return hasher(encodeLength(lhsRLP, rhsRLP))
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

func GetMerkleRoot(leaves []DualHashable) ([]byte) {
	return getMerkleRoot(util.Sha256, leaves)
}
