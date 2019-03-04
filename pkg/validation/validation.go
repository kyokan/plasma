package validation

import (
	"github.com/kyokan/plasma/pkg/chain"
	"math/big"
	"github.com/kyokan/plasma/pkg/eth"
	"github.com/kyokan/plasma/pkg/db"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/kyokan/plasma/util"
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"errors"
	"github.com/kyokan/plasma/pkg/merkle"
		)

func ValidateSpendTransaction(storage db.Storage, tx *chain.Transaction) (error) {
	if tx.Body.Output0.Amount.Cmp(big.NewInt(0)) == -1 {
		return NewErrNegativeOutput(0)
	}
	if tx.Body.Output1.Amount.Cmp(big.NewInt(0)) == -1 {
		return NewErrNegativeOutput(1)
	}

	prevTx0Conf, err := storage.FindTransactionByBlockNumTxIdx(tx.Body.Input0.BlockNumber, tx.Body.Input0.TransactionIndex)
	if err == leveldb.ErrNotFound {
		return NewErrTxNotFound(0, tx.Body.Input0.BlockNumber, tx.Body.Input0.TransactionIndex)
	}
	if err != nil {
		return err
	}
	prevTx0 := prevTx0Conf.Transaction
	if prevTx0Conf.ConfirmSigs[0] != tx.Body.Input0ConfirmSigs[0] {
		return NewErrConfirmSigMismatch(0, 0)
	}
	if prevTx0Conf.ConfirmSigs[1] != tx.Body.Input0ConfirmSigs[1] {
		return NewErrConfirmSigMismatch(0, 1)
	}
	sigHash0 := tx.Body.SignatureHash()
	prevTx0Output := prevTx0.Body.OutputAt(tx.Body.Input0.OutputIndex)
	err = eth.ValidateSignature(sigHash0, tx.Sigs[0][:], prevTx0Output.Owner)
	if err != nil {
		return NewErrInvalidSignature(0)
	}

	totalInput := big.NewInt(0)
	totalInput = totalInput.Add(totalInput, prevTx0Output.Amount)

	if !tx.Body.Input1.IsZero() {
		if tx.Body.Input1.BlockNumber == tx.Body.Input0.BlockNumber && tx.Body.Input1.TransactionIndex == tx.Body.Input0.TransactionIndex && tx.Body.Input1.OutputIndex == tx.Body.Input0.OutputIndex {
			return NewErrIdenticalInputs()
		}

		prevTx1Conf, err := storage.FindTransactionByBlockNumTxIdx(tx.Body.Input1.BlockNumber, tx.Body.Input1.TransactionIndex)
		if err == leveldb.ErrNotFound {
			return NewErrTxNotFound(1, tx.Body.Input1.BlockNumber, tx.Body.Input1.TransactionIndex)
		}
		if err != nil {
			return err
		}
		prevTx1 := prevTx1Conf.Transaction
		if prevTx1Conf.ConfirmSigs[0] != tx.Body.Input1ConfirmSigs[0] {
			return NewErrConfirmSigMismatch(1, 0)
		}
		if prevTx1Conf.ConfirmSigs[1] != tx.Body.Input1ConfirmSigs[1] {
			return NewErrConfirmSigMismatch(1, 1)
		}
		prevTx1Output := prevTx1.Body.OutputAt(tx.Body.Input1.OutputIndex)
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

	var emptySig chain.Signature
	if !tx.Body.Input1.IsZero() || tx.Body.Input1ConfirmSigs[0] != emptySig || tx.Body.Input1ConfirmSigs[1] != emptySig {
		return NewErrDepositDefinedInput1()
	}

	if tx.Body.Input0ConfirmSigs[0] != emptySig || tx.Body.Input0ConfirmSigs[1] != emptySig {
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

func ValidateConfirmSigs(storage db.Storage, client eth.Client, blk *chain.Block, confirmed *chain.ConfirmedTransaction) error {
	var emptySig chain.Signature
	tx := confirmed.Transaction
	merkleRoot := blk.Header.MerkleRoot
	txHash := tx.RLPHash(util.Sha256)
	var sigBuf bytes.Buffer
	sigBuf.Write(txHash[:])
	sigBuf.Write(merkleRoot[:])
	sigHash := util.Sha256(sigBuf.Bytes())
	for i, sig := range confirmed.ConfirmSigs {
		input := tx.Body.InputAt(uint8(i))
		if i > 0 && input.IsZero() {
			if sig != emptySig {
				return errors.New("confirmation signature must be zero for zero inputs")
			}

			continue
		}

		var err error
		var owner common.Address
		if input.IsDeposit() {
			_, owner, err = client.LookupDeposit(input.DepositNonce)
			if err != nil {
				return err
			}
		} else {
			prevTx, err := storage.FindTransactionByBlockNumTxIdx(input.BlockNumber, input.TransactionIndex)
			if err != nil {
				return err
			}
			owner = prevTx.Transaction.Body.OutputAt(input.OutputIndex).Owner
		}

		if err := eth.ValidateSignature(sigHash, sig[:], owner); err != nil {
			return err
		}
	}

	return nil
}

func ValidateConfirmedTransaction(storage db.Storage, client eth.Client, block *chain.Block, confirmed *chain.ConfirmedTransaction) error {
	var err error
	if confirmed.Transaction.Body.IsDeposit() {
		err = ValidateDepositTransaction(storage, client, confirmed.Transaction)
	} else {
		err = ValidateSpendTransaction(storage, confirmed.Transaction)
	}
	if err != nil {
		return err
	}

	return ValidateConfirmSigs(storage, client, block, confirmed)
}

func ValidateBlock(storage db.Storage, client eth.Client, block *chain.Block, confirmedTxs []chain.ConfirmedTransaction) error {
	hashables := make([]util.RLPHashable, len(confirmedTxs), len(confirmedTxs))
	for i, tx := range confirmedTxs {
		err := ValidateConfirmedTransaction(storage, client, block, &tx)
		if err != nil {
			return err
		}
		hashables[i] = tx.Transaction
	}

	merkleRoot := merkle.Root(hashables)
	if !bytes.Equal(merkleRoot, block.Header.MerkleRoot) {
		return errors.New("invalid block merkle root")
	}

	return nil
}
