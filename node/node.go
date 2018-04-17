package node

import (
	"log"
	"time"

	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/util"
)

type PlasmaNode struct {
	DB           *db.Database
	TxSink       *TransactionSink
	PlasmaClient *eth.PlasmaClient
}

func NewPlasmaNode(db *db.Database, sink *TransactionSink, plasmaClient *eth.PlasmaClient) *PlasmaNode {
	return &PlasmaNode{
		DB:           db,
		TxSink:       sink,
		PlasmaClient: plasmaClient,
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

	ticker := time.NewTicker(time.Second * 10)
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
			// TODO: this needs to be synchronized.
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
	if len(txs) == 0 {
		// skip for now
		log.Println("Skipping package blocks because there are no transactions.")
		return
	}

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

	// TODO: can we use the eth client instead.
	// Report this block to the plasma contract
	// 2018/04/17 10:22:51 Failed to submit block: Error: the tx doesn't have the correct nonce. account has nonce of: 16 tx has nonce of: 15
	// node.PlasmaClient.SubmitBlock(rlpMerkleTree(accepted))

	blockChan <- &block
}

func rlpMerkleTree(accepted []chain.Transaction) util.MerkleTree {
	hashables := make([]util.RLPHashable, len(accepted))

	for i := range accepted {
		txPtr := &accepted[i]
		hashables[i] = util.RLPHashable(txPtr)
	}

	merkle := util.TreeFromRLPItems(hashables)
	return merkle
}
