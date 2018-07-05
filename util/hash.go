package util

import (
	"github.com/ethereum/go-ethereum/crypto/sha3"
)

type Hash []byte

type Hashable interface {
	Hash() Hash
}

type RLPHashable interface {
	RLPHash() Hash
}

func DoHash(b []byte) Hash {
	hash := sha3.NewKeccak256()

	var buf []byte
	hash.Write(b)
	buf = hash.Sum(buf)

	return buf
}
