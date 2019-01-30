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
)

const MaxMempoolSize = 65534

type Mempool struct {
	txReqs      chan *txRequest
	quit        chan bool
	flushReq    chan chan []chain.ConfirmedTransaction
	txPool      []chain.ConfirmedTransaction
	depositPool []chain.ConfirmedTransaction
	storage     db.PlasmaStorage
}

type txRequest struct {
	tx  chain.ConfirmedTransaction
	res chan error
}

func NewMempool(storage db.PlasmaStorage) *Mempool {
	return &Mempool{
		txReqs:      make(chan *txRequest),
		quit:        make(chan bool),
		flushReq:    make(chan chan []chain.ConfirmedTransaction),
		txPool:      make([]chain.ConfirmedTransaction, 0),
		depositPool: make([]chain.ConfirmedTransaction, 0),
		storage:     storage,
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
				valid, err := m.VerifyTransaction(&tx)
				if err != nil {
					log.Printf("error validating transaction with hash %s: %s", tx.RLPHash(util.Sha256), err)
					req.res <- err
					continue
				}
				if !valid {
					log.Printf("transaction with hash %s is invalid: %s", tx.RLPHash(util.Sha256), err)
					req.res <- err
					continue
				}
				if tx.Transaction.IsDeposit() {
					m.depositPool = append(m.depositPool, tx)
				} else {
					m.txPool = append(m.txPool, tx)
				}

				req.res <- nil
			case resCh := <-m.flushReq:
				res := m.txPool
				m.txPool = make([]chain.ConfirmedTransaction, 0)
				resCh <- res
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

func (m *Mempool) Flush() []chain.ConfirmedTransaction {
	res := make(chan []chain.ConfirmedTransaction)
	m.flushReq <- res
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

func (m *Mempool) VerifyTransaction(tx *chain.ConfirmedTransaction) (bool, error) {
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

	return true, nil
}
