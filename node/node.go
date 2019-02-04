package node

import (
	"github.com/kyokan/plasma/chain"
		"log"
		"time"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	)

type PlasmaNode struct {
	storage db.PlasmaStorage
	txSink  *TransactionSink
	mPool   *Mempool
	client  eth.Client
}

func NewPlasmaNode(storage db.PlasmaStorage, sink *TransactionSink, mPool *Mempool, client eth.Client) *PlasmaNode {
	return &PlasmaNode{
		storage: storage,
		txSink:  sink,
		mPool:   mPool,
		client:  client,
	}
}

func (node *PlasmaNode) Start() {
	go node.awaitTxs(time.Second)
}

func (node PlasmaNode) awaitTxs(interval time.Duration) {
	log.Print("Awaiting transactions.")

	tick := time.NewTicker(interval)

	for {
		select {
		case <-tick.C:
			deposit := node.mPool.FlushDeposit()
			if deposit != nil {
				node.packageDepositBlocks(*deposit)
				continue
			}

			spends := node.mPool.FlushSpends()
			if len(spends) > 0 {
				node.packageBlock(spends)
			}
		}
	}
}

func (node PlasmaNode) packageBlock(txs []chain.ConfirmedTransaction) {
	blockResult, err := node.storage.PackageBlock(txs)
	if err != nil {
		log.Printf("Error packaging block: %s", err.Error())
		return
	}

	if blockResult != nil {
		err = node.client.SubmitBlock(blockResult.MerkleRoot, blockResult.NumberTransactions, blockResult.BlockFees, blockResult.BlockNumber)
		if err != nil {
			log.Printf("Error submiting block: %s", err.Error())
		}
	}
}

func (node PlasmaNode) packageDepositBlocks(deposit chain.ConfirmedTransaction) {
	log.Printf("packaging 1 deposit txs into block")
	depositBlock, err := node.storage.ProcessDeposit(deposit)
	if err != nil {
		log.Printf("Error packaging deposti block: %s", err.Error())
		return
	}
	err = node.client.SubmitBlock(depositBlock.MerkleRoot, depositBlock.NumberTransactions, depositBlock.BlockFees, depositBlock.BlockNumber)
	if err != nil {
		log.Printf("Error submiting deposit block: %s", err.Error())
	}

}