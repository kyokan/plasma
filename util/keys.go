package util

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io/ioutil"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pborman/uuid"
)

func ToKeyWrapper(privateKey string) *keystore.Key {
	privateKeyECDSA := ToPrivateKeyECDSA(privateKey)

	id := uuid.NewRandom()

	return &keystore.Key{
		Id:         id,
		Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		PrivateKey: privateKeyECDSA,
	}
}

func ToPrivateKeyECDSA(privateKey string) *ecdsa.PrivateKey {
	key, err := hex.DecodeString(privateKey)
	if err != nil {
		panic(err)
	}

	privateKeyECDSA, err := crypto.ToECDSA(key)

	if err != nil {
		panic(err)
	}

	return privateKeyECDSA
}

func GetPubKey(privateKeyECDSA *ecdsa.PrivateKey) common.Address {
	return crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)
}

func GetFromKeyStore(
	addr string,
	keystoreDir string,
	keystoreFile string,
	passphrase string,
) *keystore.Key {
	// Init a keystore
	ks := keystore.NewKeyStore(
		keystoreDir,
		keystore.LightScryptN,
		keystore.LightScryptP)

	// Create account definitions
	fromAccDef := accounts.Account{
		Address: common.HexToAddress(addr),
	}

	// Find the signing account
	signAcc, err := ks.Find(fromAccDef)
	if err != nil {
		fmt.Println("account keystore find error:")
		panic(err)
	}

	// Unlock the signing account
	errUnlock := ks.Unlock(signAcc, passphrase)
	if errUnlock != nil {
		fmt.Println("account unlock error:")
		panic(err)
	}

	// Open the account key file
	keyJson, readErr := ioutil.ReadFile(keystoreFile)
	if readErr != nil {
		fmt.Println("key json read error:")
		panic(readErr)
	}

	// Get the private key
	keyWrapper, keyErr := keystore.DecryptKey(keyJson, passphrase)
	if keyErr != nil {
		fmt.Println("key decrypt error:")
		panic(keyErr)
	}

	return keyWrapper
}
