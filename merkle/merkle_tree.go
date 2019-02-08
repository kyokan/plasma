package merkle

import (
	"encoding/binary"
	"github.com/kyokan/plasma/util"
	"bytes"
		"math"
		)

func getMerkleRoot(hasher util.Hasher, leaves []util.RLPHashable) ([]byte) {
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

func Root(leaves []util.RLPHashable) ([]byte) {
	return getMerkleRoot(util.Sha256, leaves)
}

func RootAndProof(leaves []util.Hash, i int64) ([]byte, []byte) {
	if len(leaves) == 0 {
		return nil, nil
	}

	if len(leaves) == 1 {
		return leaves[0], nil
	}

	pivot := int64(math.Floor(float64(len(leaves) + 1) / 2))
	var leftHash []byte
	var rightHash []byte
	var leftProof []byte
	var rightProof []byte
	var proof bytes.Buffer

	if i < pivot {
		leftHash, leftProof = RootAndProof(leaves[0:pivot], i)
		rightHash, rightProof = RootAndProof(leaves[pivot:], -1)

		if i >= 0 {
			proof.Write(leftProof)
			proof.Write(rightHash)
		}
	} else {
		leftHash, leftProof = RootAndProof(leaves[0:pivot], -1)
		rightHash, rightProof = RootAndProof(leaves[pivot:], i - pivot)

		if i >= 0 {
			proof.Write(rightProof)
			proof.Write(leftHash)
		}
	}

	return innerHash(leftHash, rightHash), proof.Bytes()
}

func innerHash(left util.Hash, right util.Hash) util.Hash {
	if len(left) != 32 {
		panic("invalid left hash length")
	}
	if len(right) != 32 {
		panic("invalid right hash length")
	}

	var buf bytes.Buffer
	buf.WriteByte(0x20)
	buf.Write(left)
	buf.WriteByte(0x20)
	buf.Write(right)
	return util.Sha256(buf.Bytes())
}