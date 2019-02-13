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
	Tx       chain.ConfirmedTransaction
	Response chan TxInclusionResponse
}

type TxInclusionResponse struct {
	MerkleRoot       util.Hash
	BlockNumber      uint64
	TransactionIndex uint32
	Error            error
}

type Mempool struct {
	txReqs          chan *txRequest
	quit            chan bool
	flushSpendReq   chan flushSpendReq
	flushDepositReq chan chan *MempoolTx
	txPool          []MempoolTx
	depositPool     []MempoolTx
	poolSpends      map[string]bool
	storage         db.PlasmaStorage
}

type txRequest struct {
	tx  chain.ConfirmedTransaction
	res chan TxInclusionResponse
}

type flushSpendReq struct {
	res  chan []MempoolTx
	done chan bool
}

func NewMempool(storage db.PlasmaStorage) *Mempool {
	return &Mempool{
		txReqs:          make(chan *txRequest),
		quit:            make(chan bool),
		flushSpendReq:   make(chan flushSpendReq),
		flushDepositReq: make(chan chan *MempoolTx),
		txPool:          make([]MempoolTx, 0),
		depositPool:     make([]MempoolTx, 0),
		poolSpends:      make(map[string]bool),
		storage:         storage,
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
				if tx.Transaction.IsDeposit() {
					err = m.VerifyDepositTransaction(&tx)
				} else {
					err = m.VerifySpendTransaction(&tx)
				}
				if err != nil {
					mPoolLogger.WithFields(logrus.Fields{
						"hash":   tx.Transaction.SignatureHash().Hex(),
						"reason": err,
					}).Warn("transaction rejected from mempool")

					req.res <- TxInclusionResponse{
						Error: err,
					}
					continue
				}
				if tx.Transaction.IsDeposit() {
					m.depositPool = append(m.depositPool, MempoolTx{
						Tx:       tx,
						Response: req.res,
					})
				} else {
					m.txPool = append(m.txPool, MempoolTx{
						Tx:       tx,
						Response: req.res,
					})
				}
				m.updatePoolSpends(&tx)
			case req := <-m.flushSpendReq:
				res := m.txPool
				m.txPool = make([]MempoolTx, 0)
				m.poolSpends = make(map[string]bool)
				req.res <- res
				<-req.done
			case resCh := <-m.flushDepositReq:
				if len(m.depositPool) == 0 {
					resCh <- nil
				} else {
					res := m.depositPool[0]
					m.depositPool = m.depositPool[1:]
					resCh <- &res
				}
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

func (m *Mempool) FlushSpends(done chan bool) []MempoolTx {
	res := make(chan []MempoolTx)
	m.flushSpendReq <- flushSpendReq{
		res:  res,
		done: done,
	}
	return <-res
}

func (m *Mempool) FlushDeposit() *MempoolTx {
	res := make(chan *MempoolTx)
	m.flushDepositReq <- res
	return <-res
}

func (m *Mempool) Append(tx chain.ConfirmedTransaction) TxInclusionResponse {
	res := make(chan TxInclusionResponse)
	req := &txRequest{
		tx:  tx,
		res: res,
	}
	m.txReqs <- req
	return <-res
}

func (m *Mempool) VerifySpendTransaction(confirmed *chain.ConfirmedTransaction) (error) {
	txLog := mPoolLogger.WithFields(logrus.Fields{
		"hash": confirmed.Transaction.SignatureHash().Hex(),
	})

	if confirmed.Transaction.Output0.Denom.Cmp(big.NewInt(0)) == -1 {
		return errors.New("transaction rejected due to negative output0 denomination")
	}

	prevTx0, err := m.storage.FindTransactionByBlockNumTxIdx(confirmed.Transaction.Input0.BlkNum, confirmed.Transaction.Input0.TxIdx)
	if err != nil {
		return err
	}
	if prevTx0 == nil {
		return errors.New("input 0 not found")
	}

	prevTx0Output := prevTx0.Transaction.OutputAt(confirmed.Transaction.Input0.OutIdx)
	sigHash0 := confirmed.Transaction.Input0.SignatureHash()
	err = eth.ValidateSignature(sigHash0, confirmed.Transaction.Sig0[:], prevTx0Output.Owner)
	if err != nil {
		txLog.Warn("transaction rejected due to invalid sig 0")
		return err
	}
	err = eth.ValidateSignature(confirmed.Transaction.SignatureHash(), confirmed.Signatures[0][:], prevTx0Output.Owner)
	if err != nil {
		txLog.Warn("transaction rejected due to invalid confirm sig 0")
		return err
	}

	totalInput := big.NewInt(0)
	totalInput = totalInput.Add(totalInput, prevTx0Output.Denom)

	if !confirmed.Transaction.Input1.IsZeroInput() {
		if confirmed.Transaction.Output1.Denom.Cmp(big.NewInt(0)) == -1 {
			return errors.New("transaction rejected due to negative output1 denomination")
		}

		prevTx1, err := m.storage.FindTransactionByBlockNumTxIdx(confirmed.Transaction.Input1.BlkNum, confirmed.Transaction.Input1.TxIdx)
		if err != nil {
			return err
		}
		if prevTx1 == nil {
			return errors.New("input 1 not found")
		}

		prevTx1Output := prevTx1.Transaction.OutputAt(confirmed.Transaction.Input1.OutIdx)
		sigHash1 := confirmed.Transaction.Input1.SignatureHash()
		err = eth.ValidateSignature(sigHash1, confirmed.Transaction.Sig1[:], prevTx1Output.Owner)
		if err != nil {
			txLog.Warn("transaction rejected due to invalid sig 1")
			return err
		}
		err = eth.ValidateSignature(confirmed.Transaction.SignatureHash(), confirmed.Signatures[1][:], prevTx1Output.Owner)
		if err != nil {
			txLog.Warn("transaction rejected due to invalid confirm sig 1")
			return err
		}

		totalInput = totalInput.Add(totalInput, prevTx1Output.Denom)
	}

	totalOutput := big.NewInt(0)
	totalOutput = totalOutput.Add(totalOutput, confirmed.Transaction.Output0.Denom)
	totalOutput = totalOutput.Add(totalOutput, confirmed.Transaction.Fee)
	if !confirmed.Transaction.Output1.IsZeroOutput() {
		totalOutput = totalOutput.Add(totalOutput, confirmed.Transaction.Output1.Denom)
	}

	if totalInput.Cmp(totalOutput) != 0 {
		txLog.Warn("transaction rejected due inputs not equalling outputs plus fees")
		return errors.New("inputs and outputs do not have the same sum")
	}

	isDoubleSpent, err := m.storage.IsDoubleSpent(confirmed)
	if err != nil {
		return err
	}

	if isDoubleSpent {
		return errors.New("transaction double spent")
	}

	return nil
}

func (m *Mempool) VerifyDepositTransaction(tx *chain.ConfirmedTransaction) (error) {
	return nil
}

func (m *Mempool) updatePoolSpends(confirmed *chain.ConfirmedTransaction) {
	tx := confirmed.Transaction
	key0 := fmt.Sprintf("%d:%d:%d", tx.Input0.BlkNum, tx.Input0.TxIdx, tx.Input0.OutIdx)
	m.poolSpends[key0] = true
	if !tx.Input1.IsZeroInput() {
		key1 := fmt.Sprintf("%d:%d:%d", tx.Input1.BlkNum, tx.Input1.TxIdx, tx.Input1.OutIdx)
		m.poolSpends[key1] = true
	}
}
