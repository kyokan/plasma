package node

import (
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/chain"
	"github.com/pkg/errors"
	"github.com/kyokan/plasma/eth"
	"bytes"
	"time"
	"github.com/kyokan/plasma/util"
	"strconv"
	"github.com/kyokan/plasma/log"
	"github.com/sirupsen/logrus"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
)

type TransactionConfirmer struct {
	storage db.PlasmaStorage
	client  eth.Client
}

var tcfLogger = log.ForSubsystem("TransactionConfirmer")

func NewTransactionConfirmer(storage db.PlasmaStorage, client eth.Client) *TransactionConfirmer {
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
		fmt.Println(blockNumber, transactionIndex)
		return nil, err
	}
	blk, err := t.storage.BlockAtHeight(blockNumber)
	if err != nil {
		fmt.Println("wtf")
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
			prevTx, err := t.storage.FindTransactionByBlockNumTxIdx(input.BlockNum, input.TxIdx)
			if err != nil {
				return nil, err
			}
			owner = prevTx.Transaction.Body.OutputAt(input.OutIdx).Owner
		}

		if err := eth.ValidateSignature(sigHash, sig[:], owner); err != nil {
			lgr.Warn("rejected confirmation due to invalid signatures")
			return nil, err
		}
	}

	lgr.Info("confirmation is valid, persisting")
	return t.storage.ConfirmTransaction(blockNumber, transactionIndex, signatures)
}

func (t *TransactionConfirmer) GetConfirmations(sig []byte, nonce uint64, blockNumber uint64, transactionIndex uint32, outIndex uint8) ([2]chain.Signature, error) {
	var sigs [2]chain.Signature

	now := uint64(time.Now().Unix())
	if nonce > now || now-nonce > 10 {
		return sigs, errors.New("invalid nonce")
	}

	confirmed, err := t.storage.FindTransactionByBlockNumTxIdx(blockNumber, transactionIndex)
	if err != nil {
		return sigs, err
	}
	tx := confirmed.Transaction
	addr := tx.Body.OutputAt(outIndex).Owner
	var buf bytes.Buffer
	buf.Write([]byte(strconv.FormatUint(nonce, 10)))
	buf.Write([]byte("kyo-plasma-mvp"))
	hash := util.Keccak256(buf.Bytes())
	if err := eth.ValidateSignature(hash[:], sig, addr); err != nil {
		return sigs, errors.New("unauthorized to view signatures")
	}

	confirmSigs, err := t.storage.ConfirmSigsFor(blockNumber, transactionIndex)
	if err != nil {
		return sigs, err
	}

	return confirmSigs, err
}
