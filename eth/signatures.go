package eth

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"errors"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/util"
)

func Sign(privKey *ecdsa.PrivateKey, hash util.Hash) (chain.Signature, error) {
	ethHash := util.GethHash(hash)
	var sig chain.Signature
	rawSig, err := crypto.Sign(ethHash, privKey)
	if err != nil {
		return sig, err
	}
	copy(sig[:], rawSig)
	return sig, nil
}

func ValidateSignature(hash, signature []byte, address common.Address) error {
	ethHash := util.GethHash(hash)

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
	if !util.AddressesEqual(&address, &signatureAddress) {
		return errors.New("invalid signature")
	}
	return nil
}
