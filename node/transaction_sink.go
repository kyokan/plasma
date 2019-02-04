package node

import (
	"errors"
		"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
		"github.com/kyokan/plasma/util"
)

type TransactionSink struct {
	c       chan chain.ConfirmedTransaction
	storage db.PlasmaStorage
}

func NewTransactionSink(storage db.PlasmaStorage) *TransactionSink {
	return &TransactionSink{c: make(chan chain.ConfirmedTransaction), storage: storage}
}

func (sink *TransactionSink) VerifyTransaction(tx *chain.ConfirmedTransaction) (bool, error) {
	inputTx1, err := sink.storage.FindTransactionByBlockNumTxIdx(tx.Transaction.Input0.BlkNum, tx.Transaction.Input0.TxIdx)
	if err != nil {
		return false, err
	}
	if inputTx1 == nil {
		return false, errors.New("input 1 not found")
	}

	inputTx2, err := sink.storage.FindTransactionByBlockNumTxIdx(tx.Transaction.Input1.BlkNum, tx.Transaction.Input1.TxIdx)
	if err != nil {
		return false, err
	}

	var prevOutput1 *chain.Output
	if tx.Transaction.Input0.OutIdx.Cmp(chain.Zero()) == 0 {
		prevOutput1 = inputTx1.Transaction.Output0
	} else {
		prevOutput1 = inputTx1.Transaction.Output1
	}

	var prevOutput2 *chain.Output
	if tx.Transaction.Input1.OutIdx.Cmp(chain.Zero()) == 0 {
		prevOutput2 = inputTx2.Transaction.Output0
	} else {
		prevOutput2 = inputTx2.Transaction.Output1
	}

	totalInput := big.NewInt(0).Add(prevOutput1.Denom, prevOutput2.Denom)
	totalOutput := big.NewInt(0).Add(tx.Transaction.Output0.Denom, tx.Transaction.Output1.Denom)
	totalOutput = totalOutput.Add(totalOutput, tx.Transaction.Fee)
	if totalInput.Cmp(totalOutput) != 0 {
		return false, errors.New("inputs and outputs do not have the same sum")
	}

	sig1Bytes, err := crypto.Ecrecover(tx.Transaction.Input0.SignatureHash(), tx.Transaction.Sig0[:])
	if err != nil {
		return false, err
	}

	sig2Bytes, err := crypto.Ecrecover(tx.Transaction.Input1.SignatureHash(), tx.Transaction.Sig1[:])
	if err != nil {
		return false, err
	}

	sig1Addr := common.BytesToAddress(sig1Bytes)
	sig2Addr := common.BytesToAddress(sig2Bytes)

	if !util.AddressesEqual(&prevOutput1.Owner, &sig1Addr) {
		return false, errors.New("input 1 signature is not valid")
	}

	if !util.AddressesEqual(&prevOutput2.Owner, &sig2Addr) {
		return false, errors.New("input 2 signature is not valid")
	}

	return true, nil
}
