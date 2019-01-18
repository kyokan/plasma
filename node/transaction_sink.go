package node

import (
	"errors"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/util"
	"github.com/kyokan/plasma/txdag"
	"github.com/kyokan/plasma/types"
)

type TransactionSink struct {
	c       chan chain.Transaction
	storage db.PlasmaStorage
}

func NewTransactionSink(storage db.PlasmaStorage) *TransactionSink {
	return &TransactionSink{c: make(chan chain.Transaction), storage: storage}
}

func (sink *TransactionSink) AcceptTransactions(ch <-chan chain.Transaction) {
	go func() {
		for {
			tx := <-ch

			valid, err := sink.VerifyTransaction(&tx)

			if !valid || err != nil {
				log.Printf("Transaction with hash %s is not valid: %s", tx.Hash(util.Sha256), err)
				continue
			}

			sink.c <- tx
		}
	}()
}

func (sink *TransactionSink) AcceptTransactionRequests(chch <-chan chan types.TransactionRequest) {
	go func() {
		for {
			ch := <-chch
			req := <-ch
			balance, err := sink.storage.Balance(&req.From)

			if err != nil {
				sendErrorResponse(ch, &req, err)
				return
			}

			if balance.Cmp(req.Amount) <= 0 {
				sendErrorResponse(ch, &req, errors.New("insufficient funds"))
				return
			}

			txs, err := sink.storage.SpendableTxs(&req.From)

			if err != nil {
				sendErrorResponse(ch, &req, errors.New("insufficient funds"))
				return
			}
			var tx *chain.Transaction
			if req.Transaction.IsZeroTransaction() {
				tx, err = txdag.FindBestUTXOs(req.From, req.To, req.Amount, txs)

				if err != nil {
					sendErrorResponse(ch, &req, err)
					return
				}
			} else {
				tx = &req.Transaction
			}

			sink.c <- *tx

			req.Response = &types.TransactionResponse{
				Transaction: tx,
			}

			ch <- req
		}
	}()
}

func (sink *TransactionSink) AcceptDepositEvents(ch <-chan eth.DepositEvent) {
	go func() {
		var deposit eth.DepositEvent
		var tx chain.Transaction
		for {
			deposit = <-ch

			tx = chain.Transaction{
				Input0: chain.ZeroInput(),
				Input1: chain.ZeroInput(),
				Output0: &chain.Output{
					Owner:        deposit.Sender,
					Denom:        deposit.Value,
					DepositNonce: deposit.DepositNonce,
				},
				Output1: chain.ZeroOutput(),
				Fee:     big.NewInt(0),
			}
			sink.c <- tx
		}
	}()
}

func (sink *TransactionSink) VerifyTransaction(tx *chain.Transaction) (bool, error) {
	inputTx1, err := sink.storage.FindTransactionByBlockNumTxIdx(tx.Input0.BlkNum, tx.Input0.TxIdx)

	if err != nil {
		return false, err
	}

	if inputTx1 == nil {
		return false, errors.New("input 1 not found")
	}

	inputTx2, err := sink.storage.FindTransactionByBlockNumTxIdx(tx.Input1.BlkNum, tx.Input1.TxIdx)

	if err != nil {
		return false, err
	}

	var prevOutput1 *chain.Output

	if tx.Input0.OutIdx.Cmp(chain.Zero()) == 0 {
		prevOutput1 = inputTx1.Output0
	} else {
		prevOutput1 = inputTx1.Output1
	}

	var prevOutput2 *chain.Output

	if tx.Input1.OutIdx.Cmp(chain.Zero()) == 0 {
		prevOutput2 = inputTx2.Output0
	} else {
		prevOutput2 = inputTx2.Output1
	}

	totalInput := big.NewInt(0).Add(prevOutput1.Denom, prevOutput2.Denom)
	totalOutput := big.NewInt(0).Add(tx.Output0.Denom, tx.Output1.Denom)
	totalOutput = totalOutput.Add(totalOutput, tx.Fee)

	if totalInput.Cmp(totalOutput) != 0 {
		return false, errors.New("inputs and outputs do not have the same sum")
	}

	sig1Bytes, err := crypto.Ecrecover(tx.SignatureHash(), tx.Sig0[:])

	if err != nil {
		return false, err
	}

	sig2Bytes, err := crypto.Ecrecover(tx.SignatureHash(), tx.Sig1[:])

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

func sendErrorResponse(ch chan<- types.TransactionRequest, req *types.TransactionRequest, err error) {
	req.Response = &types.TransactionResponse{
		Error: err,
	}

	ch <- *req
}
