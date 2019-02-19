package service

import (
	"github.com/kyokan/plasma/pkg/chain"
	"errors"
	"fmt"
	"github.com/kyokan/plasma/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/kyokan/plasma/pkg/eth"
	"github.com/kyokan/plasma/util"
	"github.com/kyokan/plasma/pkg/db"
	"github.com/kyokan/plasma/pkg/validation"
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
	storage    db.Storage
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

func NewMempool(storage db.Storage, client eth.Client) *Mempool {
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
	if err := m.ensureNoPoolSpend(tx); err != nil {
		return err
	}

	return validation.ValidateSpendTransaction(m.storage, tx)
}

func (m *Mempool) VerifyDepositTransaction(tx *chain.Transaction) error {
	if err := m.ensureNoPoolSpend(tx); err != nil {
		return err
	}

	return validation.ValidateDepositTransaction(m.storage, m.client, tx)
}

func (m *Mempool) ensureNoPoolSpend(confirmed *chain.Transaction) error {
	tx := confirmed.Body
	key0 := fmt.Sprintf("%d:%d:%d:%d", tx.Input0.BlockNum, tx.Input0.TxIdx, tx.Input0.OutIdx, tx.Input0.DepositNonce)
	_, spent := m.poolSpends[key0]
	if spent {
		return validation.NewErrDoubleSpent()
	}

	if !tx.Input1.IsZero() {
		key1 := fmt.Sprintf("%d:%d:%d:%d", tx.Input1.BlockNum, tx.Input1.TxIdx, tx.Input1.OutIdx, tx.Input1.DepositNonce)
		_, spent = m.poolSpends[key1]
		if spent {
			return validation.NewErrDoubleSpent()
		}
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
