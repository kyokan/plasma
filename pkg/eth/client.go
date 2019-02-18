package eth

import (
	"context"
	"math/big"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/kyokan/plasma/util"
	"github.com/kyokan/plasma/pkg/eth/contracts"
	"github.com/kyokan/plasma/pkg/chain"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	log2 "github.com/kyokan/plasma/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

var clientLogger = log2.ForSubsystem("EthClient")

type PlasmaClient struct {
	plasma      *contracts.Plasma
	privateKey  *ecdsa.PrivateKey
	userAddress string
}

type Exit struct {
	Owner     common.Address
	Amount    *big.Int
	BlockNum  uint64
	TxIndex   uint32
	OIndex    uint8
	StartedAt *big.Int
}

type Block struct {
	Root      []byte
	StartedAt *big.Int
}

type StartExitOpts struct {
	Transaction      chain.TransactionBody
	Input            chain.Input
	Signature        []byte
	Proof            []byte
	ConfirmSignature []byte
	CommittedFee     *big.Int
}

type ChallengeExitOpts struct {
	StartExitOpts
	ExistingInput chain.Input
}

type Client interface {
	UserAddress() common.Address
	SubmitBlock(util.Hash, uint32, *big.Int, *big.Int) error
	SubmitBlocks(merkleRoot []util.Hash, txCount []uint32, fees []*big.Int, blkNum *big.Int) error
	Deposit(amount *big.Int) (*types.Receipt, error)
	Challenge(exitingTx *chain.ConfirmedTransaction, exitingOutput uint8, exitingDepositNonce *big.Int, challengingTx *chain.ConfirmedTransaction, proof []byte) (*types.Receipt, error)

	DepositFilter(start uint64, end uint64) ([]contracts.PlasmaDeposit, uint64, error)

	ChallengedExitFilter(uint64) ([]contracts.PlasmaChallengedExit, uint64, error)

	FinalizedExitFilter(uint64) ([]contracts.PlasmaFinalizedExit, uint64, error)

	StartedTransactionExitFilter(uint64, uint64) ([]contracts.PlasmaStartedTransactionExit, uint64, error)
	StartedDepositExitFilter(uint64, uint64) ([]contracts.PlasmaStartedDepositExit, uint64, error)

	EthereumBlockHeight() (uint64, error)
	LookupDeposit(depositNonce *big.Int) (*big.Int, common.Address, error)
}

type DepositEvent struct {
	Sender       common.Address
	Value        *big.Int
	DepositNonce *big.Int
}

type clientState struct {
	client     *ethclient.Client
	rpc        *rpc.Client
	contract   *contracts.Plasma
	privateKey *ecdsa.PrivateKey
}

func NewClient(nodeUrl string, contractAddr string, privateKey *ecdsa.PrivateKey) (Client, error) {
	addr := common.HexToAddress(contractAddr)
	c, err := rpc.Dial(nodeUrl)
	if err != nil {
		return nil, err
	}

	client := ethclient.NewClient(c)
	contract, err := contracts.NewPlasma(addr, client)
	return &clientState{
		client:     client,
		rpc:        c,
		contract:   contract,
		privateKey: privateKey,
	}, nil
}

func (c *clientState) UserAddress() common.Address {
	return crypto.PubkeyToAddress(*(c.privateKey.Public()).(*ecdsa.PublicKey))
}

func (c *clientState) SubmitBlock(merkleHash util.Hash, txInBlock uint32, feesInBlock *big.Int, blkNum *big.Int) error {
	return c.SubmitBlocks([]util.Hash{merkleHash}, []uint32{txInBlock}, []*big.Int{feesInBlock}, blkNum)
}

func (c *clientState) SubmitBlocks(merkleHashes []util.Hash, txInBlocks []uint32, feesInBlocks []*big.Int, firstBlkNum *big.Int) error {
	opts := CreateKeyedTransactor(c.privateKey)
	hashes := make([][32]byte, len(merkleHashes))
	for i := 0; i < len(merkleHashes); i++ {
		copy(hashes[i][:], merkleHashes[i][:32])
	}

	bigTxInBlocks := make([]*big.Int, len(txInBlocks), len(txInBlocks))
	for i, count := range txInBlocks {
		bigTxInBlocks[i] = big.NewInt(int64(count))
	}

	_, err := ContractCall(c.client, func() (*types.Transaction, error) {
		return c.contract.SubmitBlock(opts, hashes, bigTxInBlocks, feesInBlocks, firstBlkNum)
	})
	return err
}

func (c *clientState) Deposit(amount *big.Int) (*types.Receipt, error) {
	opts := CreateKeyedTransactor(c.privateKey)
	opts.Value = amount

	clientLogger.WithFields(logrus.Fields{
		"amount":  amount.Text(10),
		"address": crypto.PubkeyToAddress(c.privateKey.PublicKey).Hex(),
	}).Info("depositing funds")

	receipt, err := ContractCall(c.client, func() (*types.Transaction, error) {
		return c.contract.Deposit(opts, crypto.PubkeyToAddress(c.privateKey.PublicKey))
	})
	if err != nil {
		return nil, err
	}

	clientLogger.WithFields(logrus.Fields{
		"amount": amount.Text(10),
		"txHash": receipt.TxHash.Hex(),
	}).Info("successfully deposited funds")

	return receipt, nil
}

func (c *clientState) Challenge(exitingTx *chain.ConfirmedTransaction, exitingOutput uint8, exitingDepositNonce *big.Int, challengingTx *chain.ConfirmedTransaction, proof []byte) (*types.Receipt, error) {
	opts := CreateKeyedTransactor(c.privateKey)

	exitingTxPos := [4]*big.Int{
		util.Uint642Big(exitingTx.Transaction.Body.BlockNumber),
		util.Uint322Big(exitingTx.Transaction.Body.TransactionIndex),
		util.Uint82Big(exitingOutput),
		exitingDepositNonce,
	}

	challengingTxPos := [2]*big.Int{
		util.Uint642Big(challengingTx.Transaction.Body.BlockNumber),
		util.Uint322Big(challengingTx.Transaction.Body.TransactionIndex),
	}

	receipt, err := ContractCall(c.client, func() (*types.Transaction, error) {
		return c.contract.ChallengeExit(opts, exitingTxPos, challengingTxPos, challengingTx.Transaction.RLP(), proof, challengingTx.ConfirmSigs[0][:])
	})
	if err != nil {
		return nil, err
	}

	clientLogger.WithFields(logrus.Fields{
		// TODO add'l fields
		"txHash": receipt.TxHash.Hex(),
	}).Info("successfully challenged exit")

	return receipt, nil
}

func (c *clientState) EthereumBlockHeight() (uint64, error) {
	header, err := c.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	return header.Number.Uint64(), nil
}

func (c *clientState) LookupDeposit(depositNonce *big.Int) (*big.Int, common.Address, error) {
	var addr common.Address
	res, err := c.contract.Deposits(&bind.CallOpts{
		Pending: false,
	}, depositNonce)
	if err != nil {
		return nil, addr, err
	}
	return res.Amount, res.Owner, nil
}
