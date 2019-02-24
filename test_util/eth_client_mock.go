package test_util

import (
	"github.com/stretchr/testify/mock"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/kyokan/plasma/pkg/chain"
	"github.com/kyokan/plasma/pkg/eth/contracts"
	"github.com/kyokan/plasma/util"
)

type EthClientMock struct {
	mock.Mock
}

func (e *EthClientMock) UserAddress() common.Address {
	panic("implement me")
}

func (e *EthClientMock) SubmitBlock(util.Hash, uint32, *big.Int, *big.Int) error {
	panic("implement me")
}

func (e *EthClientMock) SubmitBlocks(merkleRoot []util.Hash, txCount []uint32, fees []*big.Int, blkNum *big.Int) error {
	panic("implement me")
}

func (e *EthClientMock) Deposit(amount *big.Int) (*types.Receipt, error) {
	panic("implement me")
}

func (e *EthClientMock) Exit(exitingTx *chain.ConfirmedTransaction, exitingOutput uint8, proof []byte) (*types.Receipt, error) {
	panic("implement me")
}

func (e *EthClientMock) Challenge(exitingTx *chain.ConfirmedTransaction, exitingOutput uint8, exitingDepositNonce *big.Int, challengingTx *chain.ConfirmedTransaction, proof []byte) (*types.Receipt, error) {
	panic("implement me")
}

func (e *EthClientMock) DepositFilter(start uint64, end uint64) ([]contracts.PlasmaDeposit, uint64, error) {
	panic("implement me")
}

func (e *EthClientMock) ChallengedExitFilter(uint64) ([]contracts.PlasmaChallengedExit, uint64, error) {
	panic("implement me")
}

func (e *EthClientMock) FinalizedExitFilter(uint64) ([]contracts.PlasmaFinalizedExit, uint64, error) {
	panic("implement me")
}

func (e *EthClientMock) StartedTransactionExitFilter(uint64, uint64) ([]contracts.PlasmaStartedTransactionExit, uint64, error) {
	panic("implement me")
}

func (e *EthClientMock) StartedDepositExitFilter(uint64, uint64) ([]contracts.PlasmaStartedDepositExit, uint64, error) {
	panic("implement me")
}

func (e *EthClientMock) EthereumBlockHeight() (uint64, error) {
	panic("implement me")
}

func (e *EthClientMock) LookupDeposit(depositNonce *big.Int) (*big.Int, common.Address, error) {
	args := e.Called(depositNonce)
	return args.Get(0).(*big.Int), args.Get(1).(common.Address), args.Error(2)
}



