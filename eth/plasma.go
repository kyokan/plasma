package eth

import (
	"crypto/ecdsa"
	"github.com/kyokan/plasma/merkle"
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
	BlockNum  uint64
	TxIndex   uint32
	OIndex    uint8
	StartedAt *big.Int
}

type Block struct {
	Root      []byte
	StartedAt *big.Int
}

func (c *clientState) SubmitBlock(merkleHash util.Hash) error {
	opts := CreateKeyedTransactor(c.privateKey)
	var root [32]byte
	copy(root[:], merkleHash[:32])
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

	hashables := make([]merkle.DualHashable, len(opts.Txs))
	for i:= 0; i < len(opts.Txs); i++ {
		hashables[i] = &opts.Txs[i]
	}
	hashes, err := merkle.GetProof(hashables, 17, int32(opts.TxIndex))
	if err != nil {
		return err
	}
	proof := []byte{}
	for i := 0; i < len(hashes); i++ {
		proof = append(proof, hashes[i]...)
	}

	res, err := c.contract.StartExit(
		auth,
		opts.BlockNum,
		opts.TxIndex,
		opts.OutIndex,
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

	hashables := make([]merkle.DualHashable, len(opts.Txs))
	for i:= 0; i < len(opts.Txs); i++ {
		hashables[i] = &opts.Txs[i]
	}
	hashes, err := merkle.GetProof(hashables, 16, int32(opts.TxIndex))
	if err != nil {
		return err
	}
	proof := []byte{}
	for i := 0; i < len(hashes); i++ {
		proof = append(proof, hashes[i]...)
	}

	res, err := c.contract.ChallengeExit(
		auth,
		opts.ExitId,
		opts.BlockNum,
		opts.TxIndex,
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

func (c *clientState) Exit(exitId uint64) (*Exit, error) {
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

func (c *clientState) Block(blocknum uint64) (*Block, error) {
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

func (c *clientState) CurrentChildBlock() (uint64, error) {
	opts := CreateCallOpts(c.UserAddress())
	return c.contract.CurrentChildBlock(opts)
}