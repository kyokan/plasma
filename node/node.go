package node

import (
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/util"
	"log"
	"math/big"
	"time"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
)

type PlasmaNode struct {
	Storage db.PlasmaStorage
	TxSink  *TransactionSink
	Client  eth.Client
}

func NewPlasmaNode(storage db.PlasmaStorage, sink *TransactionSink, client eth.Client) *PlasmaNode {
	return &PlasmaNode{
		Storage: storage,
		TxSink:  sink,
		Client:  client,
	}
}

func (node *PlasmaNode) Start() {
	go node.awaitTxs(time.Second * 10)
}

func (node PlasmaNode) awaitTxs(interval time.Duration) {
	log.Print("Awaiting transactions.")

	tick := time.NewTicker(interval)

	for {
		select {
		case confirmed := <-node.TxSink.c:
			if confirmed.Transaction.IsDeposit() {
				log.Print("Received deposit transaction. Packaging into block.")
				tick.Stop()
				go node.packageDepositBlocks(confirmed.Transaction)
				tick = time.NewTicker(interval)
			} else {
				//signedTx, err := node.Storage.StoreTransaction(confirmed)
				node.Storage.StoreTransaction(confirmed)
			}
		case <-tick.C:
			go node.packageBlock()

		}
	}
}

func (node PlasmaNode) packageBlock() {
	blockResult, err := node.Storage.PackageCurrentBlock()
	if err != nil {
		log.Printf("Error packaging block: %s", err.Error())
		return
	}
	if blockResult != nil {
		err = node.Client.SubmitBlock(blockResult.MerkleRoot, blockResult.NumberTransactions, blockResult.BlockFees, blockResult.BlockNumber)
		if err != nil {
			log.Printf("Error submiting block: %s", err.Error())
		}
	}
}

func (node PlasmaNode) packageDepositBlocks(deposit chain.Transaction) {
	previousBlock, depositBlock, err := node.Storage.ProcessDeposit(deposit)
	if err != nil {
		log.Printf("Error packaging deposti block: %s", err.Error())
		return
	}
	if previousBlock == nil {
		err = node.Client.SubmitBlock(depositBlock.MerkleRoot, depositBlock.NumberTransactions, depositBlock.BlockFees, depositBlock.BlockNumber)
	} else {
		hashes := []util.Hash{previousBlock.MerkleRoot, depositBlock.MerkleRoot}
		txnNumbers := []*big.Int{previousBlock.NumberTransactions, depositBlock.NumberTransactions}
		fees := []*big.Int{previousBlock.BlockFees, depositBlock.BlockFees}
		err = node.Client.SubmitBlocks(hashes, txnNumbers, fees, previousBlock.BlockNumber)
	}
	if err != nil {
		log.Printf("Error submiting deposit block: %s", err.Error())
	}

}