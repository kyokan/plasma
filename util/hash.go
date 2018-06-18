package util

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto/sha3"
)

type Hash []byte

type Hashable interface {
	Hash() Hash
}

type RLPHashable interface {
	RLPHash() Hash
}

// Custom JSON unmarshal needed for test fixtures
func (h *Hash) UnmarshalJSON(data []byte) error {
	s := string(data[1:len(data)-1]) // Quotes are included in the input
	l := len(s)
	if l != len(*h) {
		*h = make([]byte, l)
	}
	copied := copy(*h, s)
	if copied != l {
		return errors.New(fmt.Sprintf("Invalid number of bytes copied for hash (%s)", s))
	}
	return nil
}

func DoHash(b []byte) Hash {
	hash := sha3.NewKeccak256()

	var buf []byte
	hash.Write(b)
	buf = hash.Sum(buf)

	return buf
}
