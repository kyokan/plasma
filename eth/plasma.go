package eth

import (
	"crypto/ecdsa"
	"log"
	"math/big"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/contracts/gen/contracts"
	"github.com/kyokan/plasma/util"
)

type PlasmaClient struct {
	plasma      *contracts.Plasma
	privateKey  *ecdsa.PrivateKey
	userAddress string
}

type Exit struct {
	Owner     common.Address
	Amount    *big.Int
	BlockNum  *big.Int
	TxIndex   *big.Int
	OIndex    *big.Int
	StartedAt *big.Int
}

type Block struct {
	Root      []byte
	StartedAt *big.Int
}

func (c *clientState) SubmitBlock(merkle *util.MerkleTree) error {
	opts := CreateKeyedTransactor(c.privateKey)
	var root [32]byte
	copy(root[:], merkle.Root.Hash[:32])
	tx, err := c.contract.SubmitBlock(opts, root)

	if err != nil {
		return err
	}

	log.Printf("Submit block pending: 0x%x\n", tx.Hash())
	return nil
}

func (c *clientState) Deposit(value *big.Int, t *chain.Transaction) error {
	opts := CreateKeyedTransactor(c.privateKey)
	opts.Value = value
	bytes, err := rlp.EncodeToBytes(&t)
	if err != nil {
		return err
	}

	tx, err := c.contract.Deposit(opts, bytes)
	if err != nil {
		return err
	}

	log.Printf("Deposit pending: 0x%x\n", tx.Hash())
	return nil
}

func (c *clientState) StartExit(opts *StartExitOpts) error {
	auth := CreateKeyedTransactor(c.privateKey)
	tx := opts.Txs[opts.TxIndex]
	bytes, err := rlp.EncodeToBytes(&tx)
	if err != nil {
		return err
	}

	merkle := CreateMerkleTree(opts.Txs)
	bigTxIdx := new(big.Int).SetUint64(uint64(opts.TxIndex))
	bigOutIdx := new(big.Int).SetUint64(uint64(opts.OutIndex))
	proof := util.CreateMerkleProof(merkle, bigTxIdx)
	res, err := c.contract.StartExit(
		auth,
		opts.BlockNum,
		bigTxIdx,
		bigOutIdx,
		bytes,
		proof,
	)

	if err != nil {
		return err
	}
	log.Printf("Start Exit pending: 0x%x\n", res.Hash())
	return nil
}

func (c *clientState) ChallengeExit(opts *ChallengeExitOpts) error {
	auth := CreateKeyedTransactor(c.privateKey)
	tx := opts.Txs[opts.TxIndex]
	bytes, err := rlp.EncodeToBytes(&tx)
	if err != nil {
		return nil
	}

	merkle := CreateMerkleTree(opts.Txs)
	bigIdx := new(big.Int).SetUint64(uint64(opts.TxIndex))
	proof := util.CreateMerkleProof(merkle, bigIdx)
	res, err := c.contract.ChallengeExit(
		auth,
		opts.ExitId,
		opts.BlockNum,
		bigIdx,
		bytes,
		proof,
	)
	if err != nil {
		return err
	}

	log.Printf("Challenge Exit pending: 0x%x\n", res.Hash())
	return nil
}

func (c *clientState) Finalize() error {
	opts := CreateKeyedTransactor(c.privateKey)
	res, err := c.contract.Finalize(opts)
	if err != nil {
		return err
	}

	log.Printf("Finalize pending: 0x%x\n", res.Hash())
	return nil
}

func (c *clientState) Exit(exitId *big.Int) (*Exit, error) {
	opts := CreateCallOpts(c.UserAddress())
	owner, amount, blocknum, txindex, oindex, startedAt, err := c.contract.GetExit(opts, exitId)
	if err != nil {
		return nil, err
	}

	return &Exit{
		owner,
		amount,
		blocknum,
		txindex,
		oindex,
		startedAt,
	}, nil
}

func (c *clientState) Block(blocknum *big.Int) (*Block, error) {
	opts := CreateCallOpts(c.UserAddress())

	log.Printf("Block for address 0x%x\n", opts.From)
	root, startedAt, err := c.contract.GetBlock(opts, blocknum)
	if err != nil {
		return nil, err
	}

	return &Block{
		root[:],
		startedAt,
	}, nil
}

func (c *clientState) CurrentChildBlock() (*big.Int, error) {
	opts := CreateCallOpts(c.UserAddress())
	return c.contract.CurrentChildBlock(opts)
}

// Note this prevents import cycle with utils.
func CreateMerkleTree(accepted []chain.Transaction) util.MerkleTree {
	hashables := make([]util.RLPHashable, len(accepted))

	for i := range accepted {
		tx := accepted[i]
		hashables[i] = util.RLPHashable(&tx)
	}

	merkle := util.TreeFromRLPItems(hashables)
	return merkle
}
