package node

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
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
		lastBlock = node.createGenesisBlock()
	}

	blockChan := make(chan *chain.Block)
	go node.awaitTxs(blockChan, time.Second * 10)
	blockChan <- lastBlock
}

func (node PlasmaNode) awaitTxs(blks chan *chain.Block, interval time.Duration) {
	log.Print("Awaiting transactions.")

	var lastBlock *chain.Block
	var mempool []chain.Transaction
	tick := time.NewTicker(interval)

	for {
		select {
		case tx := <-node.TxSink.c:
			// TODO: this needs to be synchronized.
			if tx.IsDeposit() {
				log.Print("Received deposit transaction. Packaging into block.")
				// Reset ticker, making sure it won't signal while packaging the block
				tick.Stop()
				go node.packageBlock(*lastBlock, []chain.Transaction{tx}, blks)
				tick = time.NewTicker(interval)
			} else {
				log.Print("Received regular transaction. Appending to mempool.")
				mempool = append(mempool, tx)
			}
		case block := <-blks:
			lastBlock = block
		case <-tick.C:
			go node.packageBlock(*lastBlock, mempool, blks)
			mempool = nil
		}
	}
}

func (node PlasmaNode) packageBlock(lastBlock chain.Block, txs []chain.Transaction, blockChan chan<- *chain.Block) {
	if len(txs) == 0 {
		// Skip for now because it makes logs noisy
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

	// TODO: replace previous merkle root.
	rlpMerkle := rlpMerkleTree(accepted)

	header := chain.BlockHeader{
		MerkleRoot:    merkle.Root.Hash,
		RLPMerkleRoot: rlpMerkle.Root.Hash,
		PrevHash:      lastBlock.BlockHash,
		Number:        blkNum,
	}

	block := chain.Block{
		Header:    &header,
		BlockHash: header.Hash(),
	}

	node.DB.BlockDao.Save(&block)

	blockChan <- &block

	if len(accepted) == 1 && accepted[0].IsDeposit() {
		// Skip reporting block if this is a deposit, because
		// the plasma contract already creates a plasma block on deposit
		// Submitting again here would submit duplicate deposit blocks.
		return
	}

	node.PlasmaClient.SubmitBlock(rlpMerkle)
}

func (node *PlasmaNode) createGenesisBlock() *chain.Block {
	log.Println("Creating genesis block.")

	blkNum := 1

	txs := []chain.Transaction{
		chain.Transaction{
			Input0:  chain.ZeroInput(),
			Input1:  chain.ZeroInput(),
			Sig0:    []byte{},
			Sig1:    []byte{},
			Output0: chain.ZeroOutput(),
			Output1: chain.ZeroOutput(),
			Fee:     new(big.Int),
			BlkNum:  uint64(blkNum),
			TxIdx:   0,
		},
	}

	hashables := make([]util.Hashable, len(txs))

	for i := range txs {
		txPtr := &txs[i]
		hashables[i] = util.Hashable(txPtr)
	}

	node.DB.TxDao.SaveMany(txs)
	merkle := util.TreeFromItems(hashables)
	node.DB.MerkleDao.Save(&merkle.Root)

	// TODO: replace previous merkle root.
	rlpMerkle := rlpMerkleTree(txs)

	header := chain.BlockHeader{
		MerkleRoot:    merkle.Root.Hash,
		RLPMerkleRoot: rlpMerkle.Root.Hash,
		// TODO: is it okay to omit here.
		// PrevHash:   lastBlock.BlockHash,
		Number: uint64(blkNum),
	}

	block := chain.Block{
		Header:    &header,
		BlockHash: header.Hash(),
	}

	fmt.Printf("block hash: %s\n", hex.EncodeToString(block.Header.RLPMerkleRoot))

	if err := node.DB.BlockDao.Save(&block); err != nil {
		log.Fatalf("Failed to create genesis block:%v", err)
	}

	fmt.Printf("merkle hash: %s\n", hex.EncodeToString(rlpMerkle.Root.Hash))

	// Report genesis block to plasma
	node.PlasmaClient.SubmitBlock(rlpMerkle)

	return &block
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
