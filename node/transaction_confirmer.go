package node

import (
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/chain"
	"github.com/pkg/errors"
	"github.com/kyokan/plasma/eth"
)

type TransactionConfirmer struct {
	storage db.PlasmaStorage
}

func NewTransactionConfirmer (storage db.PlasmaStorage) *TransactionConfirmer {
	return &TransactionConfirmer{
		storage: storage,
	}
}

func (t *TransactionConfirmer) Confirm(blockNumber uint64, transactionIndex uint32, signatures [2]chain.Signature) (*chain.ConfirmedTransaction, error) {
	var emptySig chain.Signature
	confirmed, err := t.storage.FindTransactionByBlockNumTxIdx(blockNumber, transactionIndex)
	if err != nil {
		return nil, err
	}

	if confirmed.Signatures[0] != emptySig {
		return nil, errors.New("transaction already confirmed")
	}

	sigHash := confirmed.Transaction.SignatureHash()
	for i, sig := range signatures {
		if sig == emptySig {
			return nil, errors.New("confirmation signature is empty")
		}

		input := confirmed.Transaction.InputAt(uint8(i))
		if i > 0 && input.IsZeroInput() {
			continue
		}

		owner := input.Owner
		if err := eth.ValidateSignature(sigHash, sig[:], owner); err != nil {
		    return nil, err
		}
	}

	return t.storage.ConfirmTransaction(blockNumber, transactionIndex, signatures)
}