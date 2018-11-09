package util

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	//"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

type Hash []byte

type Hashable interface {
	Hash() Hash
}

type RLPHashable interface {
	RLPHash() Hash
}

func DoHash(b []byte) Hash {
	hash := sha3.NewLegacyKeccak256()
	// hash := sha3.NewKeccak256()

	var buf []byte
	hash.Write(b)
	buf = hash.Sum(buf)

	return buf
}

func ValidateSignature(hash, signature []byte, address common.Address) error {
	if len(signature) == 65 && signature[64] > 26 {
		signature[64] -= 27
	}
	pubKey, err := crypto.SigToPub(hash, signature)
	if err != nil {
		return err
	}
	signatureAddress := crypto.PubkeyToAddress(*pubKey)
	if !AddressesEqual(&address, &signatureAddress) {
		return errors.New("Invalid signature")
	}
	return nil
}
