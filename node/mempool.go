package node

import (
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
				"github.com/kyokan/plasma/util"
	"errors"
	"log"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
)

const MaxMempoolSize = 65534

type Mempool struct {
	txReqs          chan *txRequest
	quit            chan bool
	flushSpendReq   chan chan []chain.ConfirmedTransaction
	flushDepositReq chan chan *chain.ConfirmedTransaction
	txPool          []chain.ConfirmedTransaction
	depositPool     []chain.ConfirmedTransaction
	poolSpends      map[string]bool
	storage         db.PlasmaStorage
}

type txRequest struct {
	tx  chain.ConfirmedTransaction
	res chan error
}

func NewMempool(storage db.PlasmaStorage) *Mempool {
	return &Mempool{
		txReqs:          make(chan *txRequest),
		quit:            make(chan bool),
		flushSpendReq:   make(chan chan []chain.ConfirmedTransaction),
		flushDepositReq: make(chan chan *chain.ConfirmedTransaction),
		txPool:          make([]chain.ConfirmedTransaction, 0),
		depositPool:     make([]chain.ConfirmedTransaction, 0),
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
					req.res <- errors.New("mempool is full")
					continue
				}

				tx := req.tx
				var valid bool
				var err error
				if tx.Transaction.IsDeposit() {
					valid, err = m.VerifyDepositTransaction(&tx)
				} else {
					valid, err = m.VerifySpendTransaction(&tx)
				}
				if err != nil {
					log.Printf("error validating transaction with hash %s: %s", hexutil.Encode(tx.RLPHash(util.Sha256)), err)
					req.res <- err
					continue
				}
				if !valid {
					log.Printf("transaction with hash %s is invalid: %s", hexutil.Encode(tx.RLPHash(util.Sha256)), err)
					req.res <- err
					continue
				}
				if tx.Transaction.IsDeposit() {
					m.depositPool = append(m.depositPool, tx)
				} else {
					m.txPool = append(m.txPool, tx)
				}
				m.updatePoolSpends(&tx)

				req.res <- nil
			case resCh := <-m.flushSpendReq:
				res := m.txPool
				m.txPool = make([]chain.ConfirmedTransaction, 0)
				m.poolSpends = make(map[string]bool)
				resCh <- res
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

func (m *Mempool) FlushSpends() []chain.ConfirmedTransaction {
	res := make(chan []chain.ConfirmedTransaction)
	m.flushSpendReq <- res
	return <-res
}

func (m *Mempool) FlushDeposit() *chain.ConfirmedTransaction {
	res := make(chan *chain.ConfirmedTransaction)
	m.flushDepositReq <- res
	return <-res
}

func (m *Mempool) Append(tx chain.ConfirmedTransaction) error {
	res := make(chan error)
	req := &txRequest{
		tx:  tx,
		res: res,
	}
	m.txReqs <- req
	return <-res
}

func (m *Mempool) VerifySpendTransaction(confirmed *chain.ConfirmedTransaction) (bool, error) {
	prevTx0, err := m.storage.FindTransactionByBlockNumTxIdx(confirmed.Transaction.Input0.BlkNum, confirmed.Transaction.Input0.TxIdx)
	if err != nil {
		return false, err
	}
	if prevTx0 == nil {
		return false, errors.New("input 0 not found")
	}

	prevTx0Output := prevTx0.Transaction.OutputAt(confirmed.Transaction.Input0.OutIdx)
	sigHash0 := confirmed.Transaction.Input0.SignatureHash()
	err = util.ValidateSignature(sigHash0, confirmed.Transaction.Sig0[:], prevTx0Output.Owner)
	if err != nil {
		panic(err)
		return false, err
	}
	err = util.ValidateSignature(confirmed.Transaction.SignatureHash(), confirmed.Signatures[0][:], prevTx0Output.Owner)
	if err != nil {
		panic(err)
		return false, err
	}

	totalInput := big.NewInt(0)
	totalInput = totalInput.Add(totalInput, prevTx0Output.Denom)

	if !confirmed.Transaction.Input1.IsZeroInput() {
		prevTx1, err := m.storage.FindTransactionByBlockNumTxIdx(confirmed.Transaction.Input1.BlkNum, confirmed.Transaction.Input1.TxIdx)
		if err != nil {
			return false, err
		}
		if prevTx1 == nil {
			return false, errors.New("input 1 not found")
		}

		prevTx1Output := prevTx1.Transaction.OutputAt(confirmed.Transaction.Input1.OutIdx)
		sigHash1 := confirmed.Transaction.Input1.SignatureHash()
		err = util.ValidateSignature(sigHash1, confirmed.Transaction.Sig1[:], prevTx1Output.Owner)
		if err != nil {
			return false, err
		}
		err = util.ValidateSignature(confirmed.Transaction.SignatureHash(), confirmed.Signatures[1][:], prevTx1Output.Owner)
		if err != nil {
			return false, err
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
		return false, errors.New("inputs and outputs do not have the same sum")
	}

	isDoubleSpent, err := m.storage.IsDoubleSpent(confirmed)
	if err != nil {
		return false, err
	}

	return !isDoubleSpent, nil
}

func (m *Mempool) VerifyDepositTransaction(tx *chain.ConfirmedTransaction) (bool, error) {
	return true, nil
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
