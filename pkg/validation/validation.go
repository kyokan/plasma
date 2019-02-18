package validation

import (
	"github.com/kyokan/plasma/pkg/chain"
	"math/big"
	"github.com/kyokan/plasma/pkg/eth"
	"github.com/kyokan/plasma/pkg/db"
	"github.com/syndtr/goleveldb/leveldb"
)

func ValidateSpendTransaction(storage db.Storage, tx *chain.Transaction) (error) {
	if tx.Body.Output0.Amount.Cmp(big.NewInt(0)) == -1 {
		return NewErrNegativeOutput(0)
	}

	prevTx0Conf, err := storage.FindTransactionByBlockNumTxIdx(tx.Body.Input0.BlockNum, tx.Body.Input0.TxIdx)
	if err == leveldb.ErrNotFound {
		return NewErrTxNotFound(0, tx.Body.Input0.BlockNum, tx.Body.Input0.TxIdx)
	}
	if err != nil {
		return err
	}
	prevTx0 := prevTx0Conf.Transaction
	if prevTx0Conf.ConfirmSigs[0] != tx.Body.Input0ConfirmSig {
		return NewErrConfirmSigMismatch(0)
	}
	sigHash0 := tx.Body.SignatureHash()
	prevTx0Output := prevTx0.Body.OutputAt(tx.Body.Input0.OutIdx)
	err = eth.ValidateSignature(sigHash0, tx.Sigs[0][:], prevTx0Output.Owner)
	if err != nil {
		return NewErrInvalidSignature(0)
	}

	totalInput := big.NewInt(0)
	totalInput = totalInput.Add(totalInput, prevTx0Output.Amount)

	if !tx.Body.Input1.IsZero() {
		if tx.Body.Output1.Amount.Cmp(big.NewInt(0)) == -1 {
			return NewErrNegativeOutput(1)
		}

		prevTx1Conf, err := storage.FindTransactionByBlockNumTxIdx(tx.Body.Input1.BlockNum, tx.Body.Input1.TxIdx)
		if err == leveldb.ErrNotFound {
			return NewErrTxNotFound(1, tx.Body.Input1.BlockNum, tx.Body.Input1.TxIdx)
		}
		if err != nil {
			return err
		}

		prevTx1 := prevTx1Conf.Transaction
		prevTx1Output := prevTx1.Body.OutputAt(tx.Body.Input1.OutIdx)
		sigHash1 := tx.Body.SignatureHash()
		err = eth.ValidateSignature(sigHash1, tx.Sigs[1][:], prevTx1Output.Owner)
		if err != nil {
			return NewErrInvalidSignature(1)
		}

		totalInput = totalInput.Add(totalInput, prevTx1Output.Amount)
	}

	totalOutput := big.NewInt(0)
	totalOutput = totalOutput.Add(totalOutput, tx.Body.Output0.Amount)
	totalOutput = totalOutput.Add(totalOutput, tx.Body.Fee)
	if !tx.Body.Output1.IsZeroOutput() {
		totalOutput = totalOutput.Add(totalOutput, tx.Body.Output1.Amount)
	}

	if totalInput.Cmp(totalOutput) != 0 {
		return NewErrInputOutputValueMismatch(totalInput, totalOutput)
	}

	isDoubleSpent, err := storage.IsDoubleSpent(tx)
	if err != nil {
		return err
	}
	if isDoubleSpent {
		return NewErrDoubleSpent()
	}

	return nil
}

func ValidateDepositTransaction(storage db.Storage, client eth.Client, tx *chain.Transaction) (error) {
	if tx.Body.Output0.Amount.Cmp(big.NewInt(0)) == -1 {
		return NewErrNegativeOutput(0)
	}
	if tx.Body.Output1.Amount.Cmp(big.NewInt(0)) == -1 {
		return NewErrNegativeOutput(1)
	}
	if !tx.Body.Input1.IsZero() {
		return NewErrDepositDefinedInput1()
	}

	var emptySig chain.Signature
	if tx.Body.Input0ConfirmSig != emptySig {
		return NewErrDepositNonEmptyConfirmSig()
	}

	total, owner, err := client.LookupDeposit(tx.Body.Input0.DepositNonce)
	if err != nil {
		return err
	}

	totalOuts := big.NewInt(0)
	totalOuts = totalOuts.Add(totalOuts, tx.Body.Output0.Amount)
	totalOuts = totalOuts.Add(totalOuts, tx.Body.Output1.Amount)
	totalOuts = totalOuts.Add(totalOuts, tx.Body.Fee)
	if total.Cmp(totalOuts) != 0 {
		return NewErrInputOutputValueMismatch(total, totalOuts)
	}

	sigHash := tx.Body.SignatureHash()
	err = eth.ValidateSignature(sigHash, tx.Sigs[0][:], owner)
	if err != nil {
		return NewErrInvalidSignature(0)
	}
	err = eth.ValidateSignature(sigHash, tx.Sigs[1][:], owner)
	if err != nil {
		return NewErrInvalidSignature(1)
		return err
	}

	isDoubleSpent, err := storage.IsDoubleSpent(tx)
	if err != nil {
		return err
	}
	if isDoubleSpent {
		return NewErrDoubleSpent()
	}

	return nil
}
