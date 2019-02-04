package node

import (
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"math/big"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/util"
	"errors"
	"log"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

func (m *Mempool) VerifySpendTransaction(tx *chain.ConfirmedTransaction) (bool, error) {
	inputTx1, err := m.storage.FindTransactionByBlockNumTxIdx(tx.Transaction.Input0.BlkNum, tx.Transaction.Input0.TxIdx)
	if err != nil {
		return false, err
	}
	if inputTx1 == nil {
		return false, errors.New("input 1 not found")
	}

	inputTx2, err := m.storage.FindTransactionByBlockNumTxIdx(tx.Transaction.Input1.BlkNum, tx.Transaction.Input1.TxIdx)
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

	isDoubleSpent, err := m.storage.IsDoubleSpent(tx)
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
	key0 := fmt.Sprintf("%d:%d:%d", tx.Input0.BlkNum.Uint64(), tx.Input0.TxIdx.Uint64(), tx.Input0.OutIdx.Uint64())
	m.poolSpends[key0] = true
	if !tx.Input1.IsZeroInput() {
		key1 := fmt.Sprintf("%d:%d:%d", tx.Input1.BlkNum.Uint64(), tx.Input1.TxIdx.Uint64(), tx.Input1.OutIdx.Uint64())
		m.poolSpends[key1] = true
	}
}
