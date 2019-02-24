package service

import (
	"github.com/kyokan/plasma/pkg/db"
	"github.com/kyokan/plasma/pkg/chain"
	"github.com/pkg/errors"
	"github.com/kyokan/plasma/pkg/eth"
	"bytes"
	"github.com/kyokan/plasma/util"
	"github.com/kyokan/plasma/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/ethereum/go-ethereum/common"
)

type TransactionConfirmer struct {
	storage db.Storage
	client  eth.Client
}

var tcfLogger = log.ForSubsystem("TransactionConfirmer")

func NewTransactionConfirmer(storage db.Storage, client eth.Client) *TransactionConfirmer {
	return &TransactionConfirmer{
		storage: storage,
		client:  client,
	}
}

func (t *TransactionConfirmer) Confirm(blockNumber uint64, transactionIndex uint32, signatures [2]chain.Signature) (*chain.ConfirmedTransaction, error) {
	lgr := tcfLogger.WithFields(logrus.Fields{
		"blockNumber":      blockNumber,
		"transactionIndex": transactionIndex,
	})

	var emptySig chain.Signature
	confirmed, err := t.storage.FindTransactionByBlockNumTxIdx(blockNumber, transactionIndex)
	if err != nil {
		return nil, err
	}
	blk, err := t.storage.BlockAtHeight(blockNumber)
	if err != nil {
		return nil, err
	}
	tx := confirmed.Transaction

	merkleRoot := blk.Header.MerkleRoot
	txHash := tx.RLPHash(util.Sha256)
	var sigBuf bytes.Buffer
	sigBuf.Write(txHash[:])
	sigBuf.Write(merkleRoot[:])
	sigHash := util.Sha256(sigBuf.Bytes())
	for i, sig := range signatures {
		if sig == emptySig {
			return nil, errors.New("confirmation signature is empty")
		}

		input := tx.Body.InputAt(uint8(i))
		if i > 0 && input.IsZero() {
			continue
		}

		var owner common.Address
		if input.IsDeposit() {
			lgr.Info("checking deposit ownership")
			_, owner, err = t.client.LookupDeposit(input.DepositNonce)
			if err != nil {
				return nil, err
			}
		} else {
			prevTx, err := t.storage.FindTransactionByBlockNumTxIdx(input.BlockNumber, input.TransactionIndex)
			if err != nil {
				return nil, err
			}
			owner = prevTx.Transaction.Body.OutputAt(input.OutputIndex).Owner
		}

		if err := eth.ValidateSignature(sigHash, sig[:], owner); err != nil {
			lgr.Warn("rejected confirmation due to invalid signatures")
			return nil, err
		}
	}

	lgr.Info("confirmation is valid, persisting")
	return t.storage.ConfirmTransaction(blockNumber, transactionIndex, signatures)
}
