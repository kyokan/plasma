package test_util

import (
	"github.com/stretchr/testify/mock"
	"github.com/kyokan/plasma/pkg/chain"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type StorageMock struct {
	mock.Mock
}

func (s *StorageMock) ProcessDeposit(tx chain.Transaction) (deposit *chain.BlockResult, err error) {
	panic("implement me")
}

func (s *StorageMock) FindTransactionsByBlockNum(blkNum uint64) ([]chain.ConfirmedTransaction, error) {
	panic("implement me")
}

func (s *StorageMock) FindTransactionByBlockNumTxIdx(blkNum uint64, txIdx uint32) (*chain.ConfirmedTransaction, error) {
	panic("implement me")
}

func (s *StorageMock) Balance(addr common.Address) (*big.Int, error) {
	panic("implement me")
}

func (s *StorageMock) SpendableTxs(addr common.Address) ([]chain.ConfirmedTransaction, error) {
	panic("implement me")
}

func (s *StorageMock) UTXOs(addr common.Address) ([]chain.ConfirmedTransaction, error) {
	panic("implement me")
}

func (s *StorageMock) BlockAtHeight(num uint64) (*chain.Block, error) {
	panic("implement me")
}

func (s *StorageMock) BlockMetaAtHeight(num uint64) (*chain.BlockMetadata, error) {
	panic("implement me")
}

func (s *StorageMock) FullBlockAtHeight(num uint64) (*chain.Block, *chain.BlockMetadata, []chain.ConfirmedTransaction, error) {
	panic("implement me")
}

func (s *StorageMock) LatestBlock() (*chain.Block, error) {
	panic("implement me")
}

func (s *StorageMock) PackageBlock(txs []chain.Transaction) (result *chain.BlockResult, err error) {
	panic("implement me")
}

func (s *StorageMock) ConfirmTransaction(blockNumber uint64, transactionIndex uint32, sigs [2]chain.Signature) (*chain.ConfirmedTransaction, error) {
	panic("implement me")
}

func (s *StorageMock) ConfirmSigsFor(blockNumber uint64, transactionIndex uint32) ([2]chain.Signature, error) {
	panic("implement me")
}

func (s *StorageMock) InsertBlock(block *chain.Block, meta *chain.BlockMetadata, txs []chain.ConfirmedTransaction) error {
	panic("implement me")
}

func (s *StorageMock) LastDepositPoll() (uint64, error) {
	panic("implement me")
}

func (s *StorageMock) SaveDepositPoll(idx uint64) error {
	panic("implement me")
}

func (s *StorageMock) LastTxExitPoll() (uint64, error) {
	panic("implement me")
}

func (s *StorageMock) SaveTxExitPoll(idx uint64) error {
	panic("implement me")
}

func (s *StorageMock) LastDepositExitPoll() (uint64, error) {
	panic("implement me")
}

func (s *StorageMock) SaveDepositExitPoll(idx uint64) error {
	panic("implement me")
}

func (s *StorageMock) MarkTransactionAsExited(plasmaBlockNum uint64, plasmaTxIdx uint32, outIdx uint8, ethBlockNumber uint64, ethTransactionHash common.Hash) error {
	panic("implement me")
}

func (s *StorageMock) IsDoubleSpent(tx *chain.Transaction) (bool, error) {
	panic("implement me")
}

func (s *StorageMock) SaveLastSubmittedBlock(num uint64) error {
	panic("implement me")
}

func (s *StorageMock) LastSubmittedBlock() (uint64, error) {
	panic("implement me")
}

func (s *StorageMock) FindDoubleSpendingTransaction(blkNum uint64, txIdx uint32, outIndex uint8) (*chain.ConfirmedTransaction, error) {
	panic("implement me")
}

func (s *StorageMock) FindDoubleSpendingDeposit(nonce *big.Int) (*chain.ConfirmedTransaction, error) {
	panic("implement me")
}



