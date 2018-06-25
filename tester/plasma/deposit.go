package plasma

import (
	"crypto/ecdsa"
	"log"

	"github.com/kyokan/plasma/util"
	"gopkg.in/urfave/cli.v1"
)

func DepositIntegrationTest(c *cli.Context) {
	contractAddress := c.GlobalString("contract-addr")
	nodeURL := c.GlobalString("node-url")
	keystoreDir := c.GlobalString("keystore-dir")
	keystoreFile := c.GlobalString("keystore-file")
	userAddress := c.GlobalString("user-address")
	privateKey := c.GlobalString("private-key")
	signPassphrase := c.GlobalString("sign-passphrase")

	var privateKeyECDSA *ecdsa.PrivateKey

	if exists(userAddress) && exists(privateKey) {
		privateKeyECDSA = util.ToPrivateKeyECDSA(privateKey)
	} else if exists(keystoreDir) &&
		exists(keystoreFile) &&
		exists(userAddress) {
		keyWrapper := util.GetFromKeyStore(userAddress, keystoreDir, keystoreFile, signPassphrase)
		privateKeyECDSA = keyWrapper.PrivateKey
	}

	if privateKeyECDSA == nil {
		log.Fatalln("Private key ecdsa not found")
	}

	plasma := CreatePlasmaClient(nodeURL, contractAddress)

	depositValue := 1000000000

	t := createDepositTx(userAddress, depositValue)
	Deposit(plasma, privateKeyECDSA, userAddress, 1000000000, &t)
}
