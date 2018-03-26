package tester

import (
	"crypto/ecdsa"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kyokan/plasma/contracts/gen/contracts"
	"github.com/urfave/cli"
)

func CreatePlasmaClient(nodeUrl string, contractAddress string) *contracts.Plasma {
	conn, err := ethclient.Dial(nodeUrl)

	if err != nil {
		log.Panic("Failed to start ETH client: ", err)
	}

	plasma, err := contracts.NewPlasma(common.HexToAddress(contractAddress), conn)

	if err != nil {
		log.Fatalf("Failed to instantiate a Token contract: %v", err)
	}

	return plasma
}

func exists(str string) bool {
	return len(str) > 0
}

func Main(c *cli.Context) {
	contractAddress := c.GlobalString("contract-addr")
	nodeURL := c.GlobalString("node-url")
	keystoreDir := c.GlobalString("keystore-dir")
	keystoreFile := c.GlobalString("keystore-file")
	userAddress := c.GlobalString("user-address")
	privateKey := c.GlobalString("private-key")
	signPassphrase := c.GlobalString("sign-passphrase")

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
		panic("Private key ecdsa not found")
	}

	plasma := CreatePlasmaClient(nodeURL, contractAddress)
	Deposit(plasma, privateKeyECDSA, 1000000000)
	time.Sleep(5 * time.Second)
	SubmitBlock(plasma, privateKeyECDSA)
	time.Sleep(5 * time.Second)
	StartExit(plasma, privateKeyECDSA)
	time.Sleep(3 * time.Second)
	DepositFilter(plasma)
	time.Sleep(3 * time.Second)
	SubmitBlockFilter(plasma)
	time.Sleep(3 * time.Second)
	ExitStartedFilter(plasma)
}
