package util

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
	"crypto/sha256"
	)

type Hash []byte

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

func ValidateSignature(hash, signature []byte, address common.Address) error {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte("\x19Ethereum Signed Message:\n32"))
	hasher.Write(hash)
	ethHash := hasher.Sum(nil)

	sigCopy := make([]byte, len(signature))
	copy(sigCopy, signature)
	if len(sigCopy) == 65 && sigCopy[64] > 26 {
		sigCopy[64] -= 27
	}

	pubKey, err := crypto.SigToPub(ethHash, sigCopy)
	if err != nil {
		return err
	}

	signatureAddress := crypto.PubkeyToAddress(*pubKey)
	if !AddressesEqual(&address, &signatureAddress) {
		return errors.New("Invalid signature")
	}
	return nil
}
