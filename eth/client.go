package eth

import (
	"context"
	"math/big"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/kyokan/plasma/util"
	"github.com/kyokan/plasma/eth/contracts"
	"github.com/kyokan/plasma/chain"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"time"
	log2 "github.com/kyokan/plasma/log"
	"github.com/sirupsen/logrus"
	"github.com/ethereum/go-ethereum/core/types"
)

const SignaturePreamble = "\x19Ethereum Signed Message:\n"

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
	Transaction      chain.Transaction
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

	DepositFilter(start uint64, end uint64) ([]contracts.PlasmaDeposit, uint64, error)

	ChallengedExitFilter(uint64) ([]contracts.PlasmaChallengedExit, uint64, error)

	FinalizedExitFilter(uint64) ([]contracts.PlasmaFinalizedExit, uint64, error)

	StartedTransactionExitFilter(uint64) ([]contracts.PlasmaStartedTransactionExit, uint64, error)
	StartedDepositExitFilter(uint64) ([]contracts.PlasmaStartedDepositExit, uint64, error)

	EthereumBlockHeight() (uint64, error)
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

	tx, err := c.contract.SubmitBlock(opts, hashes, bigTxInBlocks, feesInBlocks, firstBlkNum)
	if err != nil {
		return err
	}

	_, err = util.WithRetries(func() (interface{}, error) {
		return c.client.TransactionReceipt(context.Background(), tx.Hash())
	}, 10, 5*time.Second)
	if err != nil {
		log.Panicln("failed to submit block!", err)
	}

	return nil
}

func (c *clientState) Deposit(amount *big.Int) (*types.Receipt, error) {
	opts := CreateKeyedTransactor(c.privateKey)
	opts.Value = amount

	clientLogger.WithFields(logrus.Fields{
		"amount": amount.Text(10),
	}).Info("depositing funds")

	tx, err := c.contract.Deposit(opts, crypto.PubkeyToAddress(c.privateKey.PublicKey))
	if err != nil {
		return nil, err
	}

	rawReceipt, err := util.WithRetries(func() (interface{}, error) {
		return c.client.TransactionReceipt(context.Background(), tx.Hash())
	}, 10, 5 * time.Second)
	if err != nil {
		return nil, err
	}
	receipt := rawReceipt.(*types.Receipt)

	clientLogger.WithFields(logrus.Fields{
		"amount": amount.Text(10),
		"txHash": receipt.TxHash.Hex(),
	}).Info("successfully deposited funds")

	return receipt, nil
}

func (c *clientState) EthereumBlockHeight() (uint64, error) {
	header, err := c.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	return header.Number.Uint64(), nil
}
