package util

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pborman/uuid"
)

func CreatePrivateKeyECDSA(
	userAddress string,
	privateKey string,
	keystoreDir string,
	keystoreFile string,
	signPassphrase string,
) *ecdsa.PrivateKey {
	var privateKeyECDSA *ecdsa.PrivateKey

	if exists(userAddress) && exists(privateKey) {
		privateKeyECDSA = ToPrivateKeyECDSA(privateKey)
	} else if exists(keystoreDir) &&
		exists(keystoreFile) &&
		exists(userAddress) {
		keyWrapper := GetFromKeyStore(userAddress, keystoreDir, keystoreFile, signPassphrase)
		privateKeyECDSA = keyWrapper.PrivateKey
	}

	if privateKeyECDSA == nil {
		log.Fatalln("Private key ecdsa not found")
	}

	return privateKeyECDSA
}

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
		log.Fatalf("Failed to decode string to hex bytes: %v", err)
	}

	privateKeyECDSA, err := crypto.ToECDSA(key)

	if err != nil {
		log.Fatalf("Failed to convert key to ECDSA: %v", err)
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

func exists(str string) bool {
	return len(str) > 0
}
