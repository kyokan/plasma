package pq

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kyokan/plasma/contracts/gen/contracts"
	"github.com/kyokan/plasma/util"
)

func CreatePriorityQueueClient(nodeUrl string, contractAddress string) *contracts.PriorityQueue {
	conn, err := ethclient.Dial(nodeUrl)

	if err != nil {
		log.Panic("Failed to start ETH client: ", err)
	}

	pq, err := contracts.NewPriorityQueue(common.HexToAddress(contractAddress), conn)

	if err != nil {
		log.Fatalf("Failed to instantiate a Token contract: %v", err)
	}

	return pq
}

func Add(
	pq *contracts.PriorityQueue,
	privateKeyECDSA *ecdsa.PrivateKey,
	userAddress string,
	priority *big.Int,
) {
	auth := util.CreateAuth(privateKeyECDSA)
	tx, err := pq.Add(auth, priority)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Adding to priority queue: 0x%x\n", tx.Hash())

	time.Sleep(3 * time.Second)
}

func Remove(
	pq *contracts.PriorityQueue,
	privateKeyECDSA *ecdsa.PrivateKey,
	userAddress string,
	priority *big.Int,
) {
	auth := util.CreateAuth(privateKeyECDSA)
	tx, err := pq.Remove(auth, priority)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Remove from priority queue: 0x%x\n", tx.Hash())

	time.Sleep(3 * time.Second)
}

func Pop(
	pq *contracts.PriorityQueue,
	privateKeyECDSA *ecdsa.PrivateKey,
	userAddress string,
) {
	auth := util.CreateAuth(privateKeyECDSA)
	tx, err := pq.Pop(auth)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Pop from priority queue: 0x%x\n", tx.Hash())

	time.Sleep(3 * time.Second)
}

func GetPriorities(
	pq *contracts.PriorityQueue,
	privateKeyECDSA *ecdsa.PrivateKey,
	userAddress string,
) {
	// View needs gas even thought it's not paid for
	// opts := tester.CreateCallOpts(userAddress)
	auth := util.CreateAuth(privateKeyECDSA)

	tx, err := pq.GetPriorities(auth)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Log priorities: 0x%x\n", tx.Hash())
}
