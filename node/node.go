package node

import (
	"github.com/kyokan/plasma/chain"
	"log"
	"time"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/util"
)

type PlasmaNode struct {
	storage   db.PlasmaStorage
	mPool     *Mempool
	client    eth.Client
	submitter *BlockSubmitter
}

func NewPlasmaNode(storage db.PlasmaStorage, mPool *Mempool, client eth.Client, submitter *BlockSubmitter) *PlasmaNode {
	return &PlasmaNode{
		storage:   storage,
		mPool:     mPool,
		client:    client,
		submitter: submitter,
	}
}

func (node *PlasmaNode) Start() {
	go node.awaitTxs(100 * time.Millisecond)
}

func (node *PlasmaNode) awaitTxs(interval time.Duration) {
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

func (node *PlasmaNode) packageBlock(mtxs []MempoolTx) {
	txs := make([]chain.ConfirmedTransaction, len(mtxs), len(mtxs))
	chans := make([]chan TxInclusionResponse, len(mtxs), len(mtxs))
	for i, mtx := range mtxs {
		txs[i] = mtx.Tx
		chans[i] = mtx.Response
	}

	blockResult, err := node.storage.PackageBlock(txs)
	if err != nil {
		log.Printf("Error packaging block: %s", err.Error())
		node.notifyAwaiters(chans, nil, err)
		return
	}

	if blockResult != nil {
		node.submitter.Enqueue(*blockResult)
	}

	node.notifyAwaiters(chans, blockResult, nil)
}

func (node *PlasmaNode) notifyAwaiters(awaiters []chan TxInclusionResponse, blockRes *db.BlockResult, err error) {
	for i, awaiter := range awaiters {
		if err != nil {
			awaiter <- TxInclusionResponse{
				Error: err,
			}
		} else {
			awaiter <- TxInclusionResponse{
				MerkleRoot:       blockRes.MerkleRoot,
				BlockNumber:      util.Big2Uint64(blockRes.BlockNumber),
				TransactionIndex: uint32(i),
				Error:            err,
			}
		}
	}
}

func (node *PlasmaNode) packageDepositBlocks(depositMtx MempoolTx) {
	log.Printf("packaging 1 deposit txs into block")
	depositBlock, err := node.storage.ProcessDeposit(depositMtx.Tx)
	if err != nil {
		depositMtx.Response <- TxInclusionResponse{
			Error: err,
		}
		log.Printf("Error packaging deposti block: %s", err.Error())
		return
	}

	node.submitter.Enqueue(*depositBlock)
	depositMtx.Response <- TxInclusionResponse{
		MerkleRoot:       depositBlock.MerkleRoot,
		BlockNumber:      util.Big2Uint64(depositBlock.BlockNumber),
		TransactionIndex: 0,
		Error:            nil,
	}
}
