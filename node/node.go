package node

import (
	"log"
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
		case tx := <-node.TxSink.c:
			if tx.IsDeposit() {
				log.Print("Received deposit transaction. Packaging into block.")
				tick.Stop()
				go node.Storage.ProcessDeposit(tx)
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
	rlpMerkle, err := node.Storage.PackageCurrentBlock()
	if err != nil {
		log.Printf("Error packaging block: %s", err.Error())
		return
	}
	if rlpMerkle != nil {
		node.Client.SubmitBlock(rlpMerkle)
	}
}
