package eth

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/eth/contracts"
	"github.com/kyokan/plasma/util"
	"log"
	"math/big"
	"context"
	"time"
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

func (c *clientState) GetChildBlock(blkNum uint64) ([32]byte, *big.Int, *big.Int, *big.Int, error) {
	auth := CreateKeyedTransactor(c.privateKey)
	opts := CreateCallOpts(auth.From)
	tmp, err := c.contract.PlasmaChain(opts, big.NewInt(int64(blkNum)))
	return tmp.Header, tmp.NumTxns, tmp.FeeAmount, tmp.CreatedAt, err
}

func (c *clientState) SubmitBlock(merkleHash util.Hash, txInBlock, feesInBlock, blkNum *big.Int) error {
	return c.SubmitBlocks([]util.Hash{ merkleHash }, []*big.Int{txInBlock}, []*big.Int{ feesInBlock }, blkNum)
}

func (c *clientState) SubmitBlocks(merkleHashes []util.Hash, txInBlocks, feesInBlocks []*big.Int, firstBlkNum *big.Int) error {
	opts := CreateKeyedTransactor(c.privateKey)
	hashes := make([][32]byte, len(merkleHashes))
	for i := 0; i != len(merkleHashes); i++ {
		copy(hashes[i][:], merkleHashes[i][:32])
	}

	tx, err := c.contract.SubmitBlock(opts, hashes, txInBlocks, feesInBlocks, firstBlkNum)
	if err != nil {
		return err
	}

	log.Printf("Submit block pending: 0x%x, start block num: %d\n", tx.Hash(), firstBlkNum.Uint64())
	_, err = util.WithRetries(func() (interface{}, error) {
		return c.client.TransactionReceipt(context.Background(), tx.Hash())
	}, 10, 5*time.Second)
	if err != nil {
		log.Panicln("failed to submit block!", err)
	}

	return nil
}

func (c *clientState) Deposit(value *big.Int, t *chain.Transaction) error {
	opts := CreateKeyedTransactor(c.privateKey)
	opts.Value = value

	tx, err := c.contract.Deposit(opts, c.UserAddress())
	if err != nil {
		return err
	}

	log.Printf("Deposit pending: 0x%x\n", tx.Hash())
	return nil
}

func (c *clientState) StartDepositExit(nonce, committedFee *big.Int) error {
	auth := CreateKeyedTransactor(c.privateKey)

	res, err := c.contract.StartDepositExit(auth, nonce, committedFee)

	if err != nil {
		return err
	}
	log.Printf("Start Deposit Exit pending: 0x%x\n", res.Hash())
	return nil
}

func (c *clientState) StartTransactionExit(opts *StartExitOpts) error {
	auth := CreateKeyedTransactor(c.privateKey)
	var txPos [3]*big.Int
	txPos[0] = opts.Input.BlkNum
	txPos[1] = opts.Input.TxIdx
	txPos[2] = opts.Input.OutIdx
	encoded, err := rlp.EncodeToBytes(opts.Transaction)
	if err != nil {
		return err
	}
	res, err := c.contract.StartTransactionExit(auth, txPos, encoded, opts.Proof, opts.Signature, opts.CommittedFee)
	if err != nil {
		return err
	}
	log.Printf("Start Transaction Exit pending: 0x%x\n", res.Hash())
	return nil
}

func (c *clientState) StartFeeExit(fee *big.Int) error {
	// TODO: Implement this
	return nil
}

func (c *clientState) ChallengeExit(nonce *big.Int, opts *ChallengeExitOpts) error {
	auth := CreateKeyedTransactor(c.privateKey)
	var existingTxPos [4]*big.Int
	existingTxPos[0] = opts.ExistingInput.BlkNum
	existingTxPos[1] = opts.ExistingInput.TxIdx
	existingTxPos[2] = opts.ExistingInput.OutIdx
	existingTxPos[3] = nonce
	encoded, err := rlp.EncodeToBytes(opts.Transaction)
	if err != nil {
		return err
	}
	var challengingTxPos [2]*big.Int
	challengingTxPos[0] = opts.Transaction.BlkNum
	challengingTxPos[1] = opts.Transaction.TxIdx
	res, err := c.contract.ChallengeExit(auth, existingTxPos, challengingTxPos, encoded, opts.Proof, opts.ConfirmSignature)
	if err != nil {
		return err
	}
	log.Printf("Challenge Exit pending: 0x%x\n", res.Hash())
	return nil
}

func (c *clientState) FinalizeDepositExits() error {
	auth := CreateKeyedTransactor(c.privateKey)
	_, err := c.contract.FinalizeDepositExits(auth)
	return err
}

func (c *clientState) FinalizeTransactionExits() error {
	auth := CreateKeyedTransactor(c.privateKey)
	_, err := c.contract.FinalizeTransactionExits(auth)
	return err
}
