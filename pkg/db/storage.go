package db

import (
	"github.com/kyokan/plasma/pkg/chain"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Storage interface {
	ProcessDeposit(tx chain.Transaction) (deposit *chain.BlockResult, err error)
	FindTransactionsByBlockNum(blkNum uint64) ([]chain.ConfirmedTransaction, error)
	FindTransactionByBlockNumTxIdx(blkNum uint64, txIdx uint32) (*chain.ConfirmedTransaction, error)

	Balance(addr common.Address) (*big.Int, error)
	SpendableTxs(addr common.Address) ([]chain.ConfirmedTransaction, error)
	UTXOs(addr common.Address) ([]chain.ConfirmedTransaction, error)

	BlockAtHeight(num uint64) (*chain.Block, error)
	BlockMetaAtHeight(num uint64) (*chain.BlockMetadata, error)
	LatestBlock() (*chain.Block, error)
	PackageBlock(txs []chain.Transaction) (result *chain.BlockResult, err error)
	ConfirmTransaction(blockNumber uint64, transactionIndex uint32, sigs [2]chain.Signature) (*chain.ConfirmedTransaction, error)
	ConfirmSigsFor(blockNumber uint64, transactionIndex uint32) ([2]chain.Signature, error)
	InsertBlock(block *chain.Block, meta *chain.BlockMetadata, txs []chain.ConfirmedTransaction) error

	LastDepositPoll() (uint64, error)
	SaveDepositPoll(idx uint64) error

	LastTxExitPoll() (uint64, error)
	SaveTxExitPoll(idx uint64) error

	LastDepositExitPoll() (uint64, error)
	SaveDepositExitPoll(idx uint64) error

	MarkExitsAsSpent([]chain.Input) error

	IsDoubleSpent(tx *chain.Transaction) (bool, error)

	SaveLastSubmittedBlock(num uint64) error
	LastSubmittedBlock() (uint64, error)

	FindDoubleSpendingTransaction(blkNum uint64, txIdx uint32, outIndex uint8) (*chain.ConfirmedTransaction, error)
	FindDoubleSpendingDeposit(nonce *big.Int) (*chain.ConfirmedTransaction, error)
}
