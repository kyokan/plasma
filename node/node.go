package node

import (
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/util"
	"log"
	"time"
)

type PlasmaNode struct {
	DB     *db.Database
	TxSink *TransactionSink
}

func NewPlasmaNode(db *db.Database, sink *TransactionSink) *PlasmaNode {
	return &PlasmaNode{
		DB:     db,
		TxSink: sink,
	}
}

func (node *PlasmaNode) Start() {
	var lastBlock *chain.Block

	lastBlock, err := node.DB.BlockDao.Latest()

	if err != nil {
		log.Panic("Failed to get last block:", err)
	}

	if lastBlock == nil {
		log.Print("Creating genesis block.")

		header := &chain.BlockHeader{
			Number: 1,
		}

		lastBlock = &chain.Block{
			Header:    header,
			BlockHash: header.Hash(),
		}

		if err := node.DB.BlockDao.Save(lastBlock); err != nil {
			log.Panic("Failed to create genesis block:", err)
		}
	}

	ticker := time.NewTicker(time.Millisecond * 500)
	blockChan := make(chan *chain.Block)
	go node.awaitTxs(blockChan, ticker.C)
	blockChan <- lastBlock
}

func (node PlasmaNode) awaitTxs(blks chan *chain.Block, tick <-chan time.Time) {
	log.Print("Awaiting transactions.")

	var lastBlock *chain.Block
	var mempool []chain.Transaction

	for {
		select {
		case tx := <-node.TxSink.c:
			if tx.IsDeposit() {
				log.Print("Received deposit transaction. Packaging into block.")
				go node.packageBlock(*lastBlock, []chain.Transaction{tx}, blks)
			} else {
				log.Print("Received regular transaction. Appending to mempool.")
				mempool = append(mempool, tx)
			}
		case block := <-blks:
			lastBlock = block
		case <-tick:
			go node.packageBlock(*lastBlock, mempool, blks)
			mempool = nil
		}
	}
}

func (node PlasmaNode) packageBlock(lastBlock chain.Block, txs []chain.Transaction, blockChan chan<- *chain.Block) {
	blkNum := lastBlock.Header.Number + 1

	log.Printf("Packaging block %d containing %d transactions.", blkNum, len(txs))

	accepted, rejected := EnsureNoDoubleSpend(txs)
	hashables := make([]util.Hashable, len(accepted))

	log.Printf("Accepted %d of %d transactions. %d rejected due to double spend.",
		len(accepted), len(txs), len(rejected))

	for i := range accepted {
		txPtr := &accepted[i]
		txPtr.BlkNum = blkNum
		txPtr.TxIdx = uint32(i)
		hashables[i] = util.Hashable(txPtr)
	}

	node.DB.TxDao.SaveMany(accepted)
	merkle := util.TreeFromItems(hashables)
	node.DB.MerkleDao.Save(&merkle.Root)

	header := chain.BlockHeader{
		MerkleRoot: merkle.Root.Hash,
		PrevHash:   lastBlock.BlockHash,
		Number:     blkNum,
	}

	block := chain.Block{
		Header:    &header,
		BlockHash: header.Hash(),
	}

	node.DB.BlockDao.Save(&block)
	blockChan <- &block
}
