package node

import (
	"errors"
	"log"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/util"
)

type TransactionSink struct {
	c      chan chain.Transaction
	db     *db.Database
	client eth.Client
}

type TransactionRequest struct {
	From     common.Address
	To       common.Address
	Amount   *big.Int
	Response *TransactionResponse
}

type TransactionResponse struct {
	Error       error
	Transaction *chain.Transaction
}

func NewTransactionSink(db *db.Database, client eth.Client) *TransactionSink {
	return &TransactionSink{c: make(chan chain.Transaction), db: db, client: client}
}

func (sink *TransactionSink) AcceptTransactions(ch <-chan chain.Transaction) {
	go func() {
		for {
			tx := <-ch

			valid, err := sink.VerifyTransaction(&tx)

			if !valid || err != nil {
				log.Printf("Transaction with hash %s is not valid: %s", tx.Hash(), err)
				continue
			}

			sink.c <- tx
		}
	}()
}

func (sink *TransactionSink) AcceptTransactionRequests(chch <-chan chan TransactionRequest) {
	go func() {
		for {
			ch := <-chch
			req := <-ch
			balance, err := sink.db.AddressDao.Balance(&req.From)

			if err != nil {
				sendErrorResponse(ch, &req, err)
				return
			}

			if balance.Cmp(req.Amount) <= 0 {
				sendErrorResponse(ch, &req, errors.New("insufficient funds"))
				return
			}

			utxoTxs, err := sink.FindBestUTXOs(req.From, req.Amount)

			if err != nil {
				sendErrorResponse(ch, &req, err)
				return
			}

			var input1 *chain.Input
			var output1 *chain.Output

			if len(utxoTxs) == 1 {
				input1 = chain.ZeroInput()

				utxo := utxoTxs[0].OutputFor(&req.From)

				if utxo == nil {
					panic("expected a UTXO")
				}

				totalAmount := utxo.Amount

				output1 = &chain.Output{
					NewOwner: req.From,
					Amount:   big.NewInt(0).Sub(totalAmount, req.Amount),
				}
			} else {
				input1 = &chain.Input{
					BlkNum: utxoTxs[1].BlkNum,
					TxIdx:  utxoTxs[1].TxIdx,
					OutIdx: utxoTxs[1].OutputIndexFor(&req.From),
				}

				totalAmount := big.NewInt(0)
				totalAmount = totalAmount.Add(utxoTxs[0].OutputFor(&req.From).Amount, utxoTxs[1].OutputFor(&req.From).Amount)

				output1 = &chain.Output{
					NewOwner: req.From,
					Amount:   big.NewInt(0).Sub(totalAmount, req.Amount),
				}
			}

			tx := chain.Transaction{
				Input0: &chain.Input{
					BlkNum: utxoTxs[0].BlkNum,
					TxIdx:  utxoTxs[0].TxIdx,
					OutIdx: utxoTxs[0].OutputIndexFor(&req.From),
				},
				Input1: input1,
				Output0: &chain.Output{
					NewOwner: req.To,
					Amount:   req.Amount,
				},
				Output1: output1,
				Fee:     big.NewInt(0),
			}

			// TODO: Optionally use local private key for testing
			tx.Sig0, err = sink.client.SignData(&req.From, tx.SignatureHash())
			if err != nil {
				sendErrorResponse(ch, &req, err)
				return
			}

			tx.Sig1, err = sink.client.SignData(&req.From, tx.SignatureHash())
			if err != nil {
				sendErrorResponse(ch, &req, err)
				return
			}

			sink.c <- tx

			req.Response = &TransactionResponse{
				Transaction: &tx,
			}

			ch <- req
		}
	}()
}

func (sink *TransactionSink) AcceptDepositEvents(ch <-chan eth.DepositEvent) {
	go func() {
		for {
			deposit := <-ch

			tx := chain.Transaction{
				Input0: chain.ZeroInput(),
				Input1: chain.ZeroInput(),
				Output0: &chain.Output{
					NewOwner: deposit.Sender,
					Amount:   deposit.Value,
				},
				Output1: chain.ZeroOutput(),
				Fee:     big.NewInt(0),
			}
			sink.c <- tx
		}
	}()
}

func (sink *TransactionSink) VerifyTransaction(tx *chain.Transaction) (bool, error) {
	inputTx1, err := sink.db.TxDao.FindByBlockNumTxIdx(tx.Input0.BlkNum, tx.Input0.TxIdx)

	if err != nil {
		return false, err
	}

	if inputTx1 == nil {
		return false, errors.New("input 1 not found")
	}

	inputTx2, err := sink.db.TxDao.FindByBlockNumTxIdx(tx.Input1.BlkNum, tx.Input1.TxIdx)

	if err != nil {
		return false, err
	}

	var prevOutput1 *chain.Output

	if tx.Input0.OutIdx == 0 {
		prevOutput1 = inputTx1.Output0
	} else {
		prevOutput1 = inputTx1.Output1
	}

	var prevOutput2 *chain.Output

	if tx.Input1.OutIdx == 0 {
		prevOutput2 = inputTx2.Output0
	} else {
		prevOutput2 = inputTx2.Output1
	}

	totalInput := big.NewInt(0).Add(prevOutput1.Amount, prevOutput2.Amount)
	totalOutput := big.NewInt(0).Add(tx.Output0.Amount, tx.Output1.Amount)
	totalOutput = totalOutput.Add(totalOutput, tx.Fee)

	if totalInput.Cmp(totalOutput) != 0 {
		return false, errors.New("inputs and outputs do not have the same sum")
	}

	sig1Bytes, err := crypto.Ecrecover(tx.SignatureHash(), tx.Sig0)

	if err != nil {
		return false, err
	}

	sig2Bytes, err := crypto.Ecrecover(tx.SignatureHash(), tx.Sig1)

	if err != nil {
		return false, err
	}

	sig1Addr := common.BytesToAddress(sig1Bytes)
	sig2Addr := common.BytesToAddress(sig2Bytes)

	if !util.AddressesEqual(&prevOutput1.NewOwner, &sig1Addr) {
		return false, errors.New("input 1 signature is not valid")
	}

	if !util.AddressesEqual(&prevOutput2.NewOwner, &sig2Addr) {
		return false, errors.New("input 2 signature is not valid")
	}

	return true, nil
}

func (sink *TransactionSink) FindBestUTXOs(addr common.Address, amount *big.Int) ([]chain.Transaction, error) {
	txs, err := sink.db.AddressDao.SpendableTxs(&addr)

	if err != nil {
		return nil, err
	}

	// similar algo to the one Bitcoin uses: https://bitcoin.stackexchange.com/questions/1077/what-is-the-coin-selection-algorithm

	// pass 1: find UTXOs that exactly match amount
	for _, tx := range txs {
		utxo := tx.OutputFor(&addr)

		if utxo.Amount.Cmp(amount) == 0 {
			return []chain.Transaction{tx}, nil
		}
	}

	// pass 2: sum of any two UTXOs exactly matches amount
	for i, ltx := range txs {
		for _, rtx := range txs[i:] {
			total := big.NewInt(0)
			total = total.Add(ltx.OutputFor(&addr).Amount, rtx.OutputFor(&addr).Amount)

			if total.Cmp(amount) == 0 {
				return []chain.Transaction{ltx, rtx}, nil
			}
		}
	}

	// pass 3: find smallest UTXO larger than amount
	var ret *chain.Transaction
	var closestAmount *big.Int

	for _, tx := range txs {
		utxo := tx.OutputFor(&addr)

		if utxo.Amount.Cmp(amount) == -1 {
			continue
		}

		if closestAmount == nil {
			closestAmount = utxo.Amount
			ret = &tx
		}

		if closestAmount.Cmp(utxo.Amount) == -1 {
			continue
		}

		closestAmount = utxo.Amount
		ret = &tx
	}

	if ret != nil {
		return []chain.Transaction{*ret}, nil
	}

	// pass 4: randomly permute utxos until sum is greater than amount or cap is reached
	for i := 0; i < 1000; i++ {
		lIdx := rand.Intn(len(txs))
		rIdx := lIdx

		for lIdx == rIdx {
			rIdx = rand.Intn(len(txs))
		}

		ltx := txs[lIdx]
		rtx := txs[rIdx]

		total := big.NewInt(0).Add(ltx.OutputFor(&addr).Amount, rtx.OutputFor(&addr).Amount)

		if total.Cmp(amount) == 1 {
			return []chain.Transaction{ltx, rtx}, nil
		}
	}

	return nil, errors.New("no suitable UTXOs found")
}

func sendErrorResponse(ch chan<- TransactionRequest, req *TransactionRequest, err error) {
	req.Response = &TransactionResponse{
		Error: err,
	}

	ch <- *req
}
