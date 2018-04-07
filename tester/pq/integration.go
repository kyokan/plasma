package pq

import (
	"crypto/ecdsa"
	"time"

	"github.com/kyokan/plasma/contracts/gen/contracts"
	"github.com/kyokan/plasma/util"
	"github.com/urfave/cli"
)

func IntegrationTest(c *cli.Context) {
	priorityQueueContractAddress := c.GlobalString("priority-queue-contract-addr")
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
		panic("Private key ecdsa not found")
	}

	pq := CreatePriorityQueueClient(nodeURL, priorityQueueContractAddress)

	addTest(pq, privateKeyECDSA, userAddress)
	removeTest(pq, privateKeyECDSA, userAddress)
	bubbleDownTest(pq, privateKeyECDSA, userAddress)
	pruneTest(pq, privateKeyECDSA, userAddress)
	popTest(pq, privateKeyECDSA, userAddress)
	logs(pq)
}

func addTest(
	pq *contracts.PriorityQueue,
	privateKeyECDSA *ecdsa.PrivateKey,
	userAddress string,
) {
	Add(pq, privateKeyECDSA, userAddress, util.NewInt(3000000000))
	Add(pq, privateKeyECDSA, userAddress, util.NewInt(2000010000))
	Add(pq, privateKeyECDSA, userAddress, util.NewInt(1000010000))

	GetPriorities(pq, privateKeyECDSA, userAddress)
}

func removeTest(
	pq *contracts.PriorityQueue,
	privateKeyECDSA *ecdsa.PrivateKey,
	userAddress string,
) {
	Remove(pq, privateKeyECDSA, userAddress, util.NewInt(1000010000))
	Remove(pq, privateKeyECDSA, userAddress, util.NewInt(2000010000))
	Remove(pq, privateKeyECDSA, userAddress, util.NewInt(3000000000))

	GetPriorities(pq, privateKeyECDSA, userAddress)
}

func bubbleDownTest(
	pq *contracts.PriorityQueue,
	privateKeyECDSA *ecdsa.PrivateKey,
	userAddress string,
) {
	Add(pq, privateKeyECDSA, userAddress, util.NewInt(3000000000))
	Add(pq, privateKeyECDSA, userAddress, util.NewInt(2000010000))
	Add(pq, privateKeyECDSA, userAddress, util.NewInt(1000010000))
	Add(pq, privateKeyECDSA, userAddress, util.NewInt(4000000000))
	Add(pq, privateKeyECDSA, userAddress, util.NewInt(5000000000))
	Add(pq, privateKeyECDSA, userAddress, util.NewInt(6000000000))
	Add(pq, privateKeyECDSA, userAddress, util.NewInt(7000000000))

	GetPriorities(pq, privateKeyECDSA, userAddress)

	Remove(pq, privateKeyECDSA, userAddress, util.NewInt(1000010000))

	GetPriorities(pq, privateKeyECDSA, userAddress)

	Remove(pq, privateKeyECDSA, userAddress, util.NewInt(2000010000))

	GetPriorities(pq, privateKeyECDSA, userAddress)
}

func pruneTest(
	pq *contracts.PriorityQueue,
	privateKeyECDSA *ecdsa.PrivateKey,
	userAddress string,
) {
	Remove(pq, privateKeyECDSA, userAddress, util.NewInt(4000000000))
	Remove(pq, privateKeyECDSA, userAddress, util.NewInt(5000000000))
	Remove(pq, privateKeyECDSA, userAddress, util.NewInt(7000000000))

	GetPriorities(pq, privateKeyECDSA, userAddress)
}

func popTest(
	pq *contracts.PriorityQueue,
	privateKeyECDSA *ecdsa.PrivateKey,
	userAddress string,
) {
	Pop(pq, privateKeyECDSA, userAddress)
	Pop(pq, privateKeyECDSA, userAddress)
	Pop(pq, privateKeyECDSA, userAddress)

	GetPriorities(pq, privateKeyECDSA, userAddress)
}

func logs(pq *contracts.PriorityQueue) {
	DebugUintFilter(pq)
	time.Sleep(3 * time.Second)
}

func exists(str string) bool {
	return len(str) > 0
}
