package node

import (
	"log"
	"time"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
)

type PlasmaNode struct {
	Storage      db.PlasmaStorage
	TxSink       *TransactionSink
	PlasmaClient *eth.PlasmaClient
}

func NewPlasmaNode(storage db.PlasmaStorage, sink *TransactionSink, plasmaClient *eth.PlasmaClient) *PlasmaNode {
	return &PlasmaNode{
		Storage:      storage,
		TxSink:       sink,
		PlasmaClient: plasmaClient,
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
		case tx := <-node.TxSink.c:
			if tx.IsDeposit() {
				log.Print("Received deposit transaction. Packaging into block.")
				tick.Stop()
				prev, _, _ := node.Storage.ProcessDeposit(tx)
				node.PlasmaClient.SubmitBlock(*prev)
				tick = time.NewTicker(interval)
			} else {
				node.Storage.StoreTransaction(tx)
			}
		case <-tick.C:
			go node.packageBlock()

		}
	}
}

func (node PlasmaNode) packageBlock() {
	rlpMerkle, _ := node.Storage.PackageCurrentBlock()
	node.PlasmaClient.SubmitBlock(*rlpMerkle)
}
