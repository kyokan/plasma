package util

import (
	"golang.org/x/crypto/sha3"
	"crypto/sha256"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"encoding/json"
)

type Hash []byte

func (h Hash) Hex() string {
	return hexutil.Encode(h)
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.Hex())
}

func (h *Hash) UnmarshalJSON(in []byte) error {
	var hashStr string
	err := json.Unmarshal(in, &hashStr)
	if err != nil {
		return err
	}
	res, err := hexutil.Decode(hashStr)
	if err != nil {
		return err
	}
	var hash [32]byte
	copy(hash[:], res)
	*h = hash[:]
	return nil
}

type Hasher func([]byte) Hash

type Hashable interface {
	Hash(Hasher) Hash
}

type RLPHashable interface {
	RLPHash(Hasher) Hash
}

func Keccak256(b []byte) Hash {
	hash := sha3.NewLegacyKeccak256()

	var buf []byte
	hash.Write(b)
	buf = hash.Sum(buf)

	return buf
}

func Sha256(b []byte) Hash {
	hash := sha256.Sum256(b)
	return hash[:]
}

func GethHash(b []byte) Hash {
	if len(b) != 32 {
		panic("hash must be 32 bytes")
	}

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte("\x19Ethereum Signed Message:\n32"))
	hasher.Write(b)
	return hasher.Sum(nil)
}
