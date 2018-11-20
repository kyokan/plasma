package eth

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/plasma-mvp-rootchain/gen/contracts"
	"github.com/kyokan/plasma/util"
	"log"
	"math/big"
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

	tx, err := c.contract.Deposit(opts, c.UserAddress())
	if err != nil {
		return err
	}

	log.Printf("Deposit pending: 0x%x\n", tx.Hash())
	return nil
}


func (c *clientState) GetChildBlock(nonce *big.Int) ([32]byte, *big.Int, error) {
	opts := CreateCallOpts(c.UserAddress())
	return c.contract.GetChildBlock(opts, nonce)
}

func (c *clientState) StartDepositExit(nonce *big.Int) error {
	auth := CreateKeyedTransactor(c.privateKey)

	res, err := c.contract.StartDepositExit(auth, nonce)

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
	res, err := c.contract.StartTransactionExit(auth, txPos, encoded, opts.Proof, opts.Signature, opts.ConfirmSignature)
	if err != nil {
		return err
	}
	log.Printf("Start Transaction Exit pending: 0x%x\n", res.Hash())
	return nil
}

func (c *clientState) ChallengeDepositExit(nonce *big.Int, opts *ChallengeExitOpts) error {
	auth := CreateKeyedTransactor(c.privateKey)
	var existingTxPos [3]*big.Int
	existingTxPos[0] = opts.ExistingInput.BlkNum
	existingTxPos[1] = opts.ExistingInput.TxIdx
	existingTxPos[2] = opts.ExistingInput.OutIdx
	encoded, err := rlp.EncodeToBytes(opts.Transaction)
	if err != nil {
		return err
	}
	res, err := c.contract.ChallengeDepositExit(auth, nonce, existingTxPos, encoded, opts.Signature, opts.Proof, opts.ConfirmSignature)
	if err != nil {
		return err
	}
	log.Printf("Challenge Deposit Exit pending: 0x%x\n", res.Hash())
	return nil
}

func (c *clientState) ChallengeTransactionExit(opts *ChallengeExitOpts) error {
	auth := CreateKeyedTransactor(c.privateKey)
	var existingTxPos, txPos [3]*big.Int
	existingTxPos[0] = opts.ExistingInput.BlkNum
	existingTxPos[1] = opts.ExistingInput.TxIdx
	existingTxPos[2] = opts.ExistingInput.OutIdx
	txPos[0] = opts.Input.BlkNum
	txPos[1] = opts.Input.TxIdx
	txPos[2] = opts.Input.OutIdx
	encoded, err := rlp.EncodeToBytes(opts.Transaction)
	if err != nil {
		return err
	}
	res, err := c.contract.ChallengeTransactionExit(auth, existingTxPos, txPos, encoded, opts.Signature, opts.Proof, opts.ConfirmSignature)
	if err != nil {
		return err
	}
	log.Printf("Challenge Transaction Exit pending: 0x%x\n", res.Hash())
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
