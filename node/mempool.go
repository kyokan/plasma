package node

import (
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"errors"
	"fmt"
	"math/big"
	"github.com/kyokan/plasma/log"
	"github.com/sirupsen/logrus"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/util"
)

const MaxMempoolSize = 65534

var mPoolLogger = log.ForSubsystem("Mempool")

type MempoolTx struct {
	Tx       chain.Transaction
	Response chan TxInclusionResponse
}

type TxInclusionResponse struct {
	MerkleRoot       util.Hash
	BlockNumber      uint64
	TransactionIndex uint32
	Error            error
}

type Mempool struct {
	txReqs     chan *txRequest
	quit       chan bool
	flushReq   chan flushSpendReq
	txPool     []MempoolTx
	poolSpends map[string]bool
	storage    db.PlasmaStorage
	client     eth.Client
}

type txRequest struct {
	tx  chain.Transaction
	res chan TxInclusionResponse
}

type flushSpendReq struct {
	res  chan []MempoolTx
	done chan bool
}

func NewMempool(storage db.PlasmaStorage, client eth.Client) *Mempool {
	return &Mempool{
		txReqs:     make(chan *txRequest),
		quit:       make(chan bool),
		flushReq:   make(chan flushSpendReq),
		txPool:     make([]MempoolTx, 0),
		poolSpends: make(map[string]bool),
		storage:    storage,
		client:     client,
	}
}

func (m *Mempool) Start() error {
	go func() {
		for {
			select {
			case req := <-m.txReqs:
				if len(m.txPool) == MaxMempoolSize {
					req.res <- TxInclusionResponse{
						Error: errors.New("mempool is full"),
					}
					continue
				}

				tx := req.tx
				var err error
				if tx.Body.IsDeposit() {
					err = m.VerifyDepositTransaction(&tx)
				} else {
					err = m.VerifySpendTransaction(&tx)
				}
				if err != nil {
					mPoolLogger.WithFields(logrus.Fields{
						"hash":   tx.Body.SignatureHash().Hex(),
						"reason": err,
					}).Warn("transaction rejected from mempool")

					req.res <- TxInclusionResponse{
						Error: err,
					}
					continue
				}
				m.txPool = append(m.txPool, MempoolTx{
					Tx:       tx,
					Response: req.res,
				})
				m.updatePoolSpends(&tx)
			case req := <-m.flushReq:
				res := m.txPool
				m.txPool = make([]MempoolTx, 0)
				m.poolSpends = make(map[string]bool)
				req.res <- res
				<-req.done
			case <-m.quit:
				return
			}
		}
	}()
	return nil
}

func (m *Mempool) Stop() error {
	m.quit <- true
	return nil
}

func (m *Mempool) Flush(done chan bool) []MempoolTx {
	res := make(chan []MempoolTx)
	m.flushReq <- flushSpendReq{
		res:  res,
		done: done,
	}
	return <-res
}

func (m *Mempool) Append(tx chain.Transaction) TxInclusionResponse {
	res := make(chan TxInclusionResponse)
	req := &txRequest{
		tx:  tx,
		res: res,
	}
	m.txReqs <- req
	return <-res
}

func (m *Mempool) VerifySpendTransaction(tx *chain.Transaction) (error) {
	txLog := mPoolLogger.WithFields(logrus.Fields{
		"hash": tx.Body.SignatureHash().Hex(),
	})
	txLog.Debug("verifying spend")

	if tx.Body.Output0.Amount.Cmp(big.NewInt(0)) == -1 {
		return errors.New("transaction rejected due to negative output0 denomination")
	}

	prevTx0Conf, err := m.storage.FindTransactionByBlockNumTxIdx(tx.Body.Input0.BlockNum, tx.Body.Input0.TxIdx)
	if err != nil {
		fmt.Println(tx.Body.Input0.BlockNum, tx.Body.Input0.TxIdx)
		return err
	}
	if prevTx0Conf == nil {
		return errors.New("input 0 not found")
	}
	prevTx0 := prevTx0Conf.Transaction
	if prevTx0Conf.ConfirmSigs[0] != tx.Body.Input0ConfirmSig {
		return errors.New("mismatched confirm sig on input0")
	}
	sigHash0 := tx.Body.SignatureHash()
	prevTx0Output := prevTx0.Body.OutputAt(tx.Body.Input0.OutIdx)
	err = eth.ValidateSignature(sigHash0, tx.Sigs[0][:], prevTx0Output.Owner)
	if err != nil {
		txLog.Warn("transaction rejected due to invalid sig 0")
		return err
	}

	totalInput := big.NewInt(0)
	totalInput = totalInput.Add(totalInput, prevTx0Output.Amount)

	if !tx.Body.Input1.IsZero() {
		if tx.Body.Output1.Amount.Cmp(big.NewInt(0)) == -1 {
			return errors.New("transaction rejected due to negative output1 denomination")
		}

		prevTx1Conf, err := m.storage.FindTransactionByBlockNumTxIdx(tx.Body.Input1.BlockNum, tx.Body.Input1.TxIdx)
		if err != nil {
			return err
		}
		if prevTx1Conf == nil {
			return errors.New("input 1 not found")
		}
		prevTx1 := prevTx1Conf.Transaction

		prevTx1Output := prevTx1.Body.OutputAt(tx.Body.Input1.OutIdx)
		sigHash1 := tx.Body.SignatureHash()
		err = eth.ValidateSignature(sigHash1, tx.Sigs[1][:], prevTx1Output.Owner)
		if err != nil {
			txLog.Warn("transaction rejected due to invalid sig 1")
			return err
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
		txLog.Warn("transaction rejected due inputs not equalling outputs plus fees")
		return errors.New("inputs and outputs do not have the same sum")
	}

	isDoubleSpent, err := m.storage.IsDoubleSpent(tx)
	if err != nil {
		return err
	}

	if isDoubleSpent {
		return errors.New("transaction double spent")
	}

	return nil
}

func (m *Mempool) VerifyDepositTransaction(tx *chain.Transaction) (error) {
	txLog := mPoolLogger.WithFields(logrus.Fields{
		"hash": tx.Body.SignatureHash().Hex(),
	})
	txLog.Debug("verifying deposit spend")

	if tx.Body.Output0.Amount.Cmp(big.NewInt(0)) == -1 {
		txLog.Warn("deposit spend rejected due to negative output0")
		return errors.New("deposit spend rejected due to negative output0 denomination")
	}
	if tx.Body.Output1.Amount.Cmp(big.NewInt(0)) == -1 {
		txLog.Warn("deposit spend rejected due to negative output1")
		return errors.New("deposit spend rejected due to negative output1 denomination")
	}
	if !tx.Body.Input1.IsZero() {
		txLog.Warn("deposit spend rejected due to presence of non-zero input1")
		return errors.New("deposit spends cannot define input1")
	}

	var emptySig chain.Signature
	if tx.Body.Input0ConfirmSig != emptySig {
		txLog.Warn("deposit spend rejected due to non-empty input0 confirm sig")
		return errors.New("confirm sig for input 0 must be empty")
	}

	total, owner, err := m.client.LookupDeposit(tx.Body.Input0.DepositNonce)
	if err != nil {
		return err
	}

	totalOuts := big.NewInt(0)
	totalOuts = totalOuts.Add(totalOuts, tx.Body.Output0.Amount)
	totalOuts = totalOuts.Add(totalOuts, tx.Body.Output1.Amount)
	totalOuts = totalOuts.Add(totalOuts, tx.Body.Fee)
	if total.Cmp(totalOuts) != 0 {
		txLog.WithFields(logrus.Fields{
			"expectedSum": total.Text(10),
			"receivedSum": totalOuts.Text(10),
		}).Warn("deposit spend rejected due to inputs and outputs having different sum")
		return errors.New("inputs and outputs do not have the same sum")
	}

	sigHash := tx.Body.SignatureHash()
	err = eth.ValidateSignature(sigHash, tx.Sigs[0][:], owner)
	if err != nil {
		txLog.Warn("deposit spend rejected due to invalid sig0")
		return err
	}
	err = eth.ValidateSignature(sigHash, tx.Sigs[1][:], owner)
	if err != nil {
		txLog.Warn("deposit spend rejected due to invalid sig1")
		return err
	}

	isDoubleSpent, err := m.storage.IsDoubleSpent(tx)
	if err != nil {
		return err
	}
	if isDoubleSpent {
		txLog.Warn("found double spend for deposit spend")
		return errors.New("transaction double spent")
	}

	return nil
}

func (m *Mempool) updatePoolSpends(confirmed *chain.Transaction) {
	tx := confirmed.Body
	key0 := fmt.Sprintf("%d:%d:%d:%d", tx.Input0.BlockNum, tx.Input0.TxIdx, tx.Input0.OutIdx, tx.Input0.DepositNonce)
	m.poolSpends[key0] = true
	if !tx.Input1.IsZero() {
		key1 := fmt.Sprintf("%d:%d:%d:%d", tx.Input1.BlockNum, tx.Input1.TxIdx, tx.Input1.OutIdx, tx.Input1.DepositNonce)
		m.poolSpends[key1] = true
	}
}
