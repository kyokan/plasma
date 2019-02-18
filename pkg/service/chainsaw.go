package service

import (
	"time"
	"github.com/kyokan/plasma/pkg/eth"
	"github.com/kyokan/plasma/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/kyokan/plasma/pkg/chain"
	"math/big"
	"github.com/kyokan/plasma/pkg/db"
	"sync"
	"github.com/kyokan/plasma/util"
	"github.com/kyokan/plasma/pkg/merkle"
)

var logger = log.ForSubsystem("Chainsaw")

type Chainsaw struct {
	quit    chan bool
	client  eth.Client
	mPool   *Mempool
	storage db.Storage
}

func NewChainsaw(client eth.Client, mPool *Mempool, storage db.Storage) *Chainsaw {
	return &Chainsaw{
		quit:    make(chan bool),
		client:  client,
		mPool:   mPool,
		storage: storage,
	}
}

func (c *Chainsaw) Start() error {
	go func() {
		logger.Info("chainsaw started")

		for {
			select {
			case <-c.quit:
				return
			default:
				now := time.Now()
				c.poll()
				duration := time.Since(now)
				if duration < 5*time.Second {
					time.Sleep((5 * time.Second) - duration)
				}
			}
		}
	}()

	return nil
}

func (c *Chainsaw) Stop() error {
	c.quit <- true
	return nil
}

func (c *Chainsaw) poll() {
	head, err := c.client.EthereumBlockHeight()
	if err != nil {
		log.WithError(logger, err).Error("failed to fetch Ethereum block height")
		return
	}

	logger.Info("processing blocks")

	var wg sync.WaitGroup
	wg.Add(2)
	go c.processTxExits(&wg, head)
	go c.processDepositExits(&wg, head)
	wg.Wait()
}

func (c *Chainsaw) processTxExits(wg *sync.WaitGroup, head uint64) {
	var err error

	tail, err := c.storage.LastTxExitPoll()
	if err != nil {
		wg.Done()
		log.WithError(logger, err).Error("failed to fetch last seen block in transaction exit poller")
		return
	}
	tail += 1

	logFields := logger.WithFields(logrus.Fields{
		"head":            head,
		"tail":            tail,
		"chainsawProcess": "txExits",
	})

	defer func() {
		if err != nil {
			wg.Done()
			return
		}

		if err = c.storage.SaveTxExitPoll(head); err != nil {
			log.WithError(logFields, err).Error("failed to persist exit poll")
		}

		wg.Done()
	}()

	if tail > head {
		logFields.Warn("head is behind last block, implies bug")
		return
	}

	events, _, err := c.client.StartedTransactionExitFilter(tail, head)
	if err != nil {
		log.WithError(logFields, err).Error("failed to filter transaction exits")
	}
	if len(events) == 0 {
		logFields.Info("no transaction exits found")
		return
	}

	logFields.WithFields(logrus.Fields{
		"exitCount": len(events),
	}).Info("found transaction exits, checking")

	//TODO: check for race here between incoming transactions, potentially by requerying older blocks
	for _, event := range events {
		position := event.Position
		blkNum := util.Big2Uint64(position[0])
		txIdx := util.Big2Uint32(position[1])
		outIdx := util.Big2Uint8(position[2])
		evFields := logrus.Fields{
			"blockNumber":      blkNum,
			"transactionIndex": txIdx,
			"outputIndex":      outIdx,
			"amount":           event.Amount.Text(10),
			"owner":            event.Owner.Hex(),
		}
		challengingTx, err := c.storage.FindDoubleSpendingTransaction(blkNum, txIdx, outIdx)
		if err != nil {
			log.WithError(logFields, err).WithFields(evFields).Error("failed to query double spends")
			return
		}
		if challengingTx == nil {
			logFields.WithFields(evFields).Info("transaction is not double spent")
			continue
		}
		logFields.WithFields(evFields).Info("found double spend, generating proof")
		exitingTx, err := c.storage.FindTransactionByBlockNumTxIdx(blkNum, txIdx)
		if err != nil {
			log.WithError(logFields, err).WithFields(evFields).Error("failed to query exiting transaction")
			return
		}

		txsInChallengeBlock, err := c.storage.FindTransactionsByBlockNum(challengingTx.Transaction.Body.BlockNumber)
		if err != nil {
			log.WithError(logFields, err).WithFields(evFields).Error("failed to query transactions in block")
			return
		}

		// note: proof is expected to be nil for single-transaction proofs
		proof := genTxMerkleProof(txsInChallengeBlock, txIdx)
		_, err = c.client.Challenge(exitingTx, outIdx, big.NewInt(0), challengingTx, proof)
		if err != nil {
			log.WithError(logFields, err).WithFields(evFields).Error("failed to broadcast exit challenge")
		}
	}
}

func (c *Chainsaw) processDepositExits(wg *sync.WaitGroup, head uint64) {
	var err error

	tail, err := c.storage.LastDepositExitPoll()
	if err != nil {
		wg.Done()
		log.WithError(logger, err).Error("failed to fetch last seen block in deposit exit poller")
		return
	}
	tail += 1

	logFields := logger.WithFields(logrus.Fields{
		"head":            head,
		"tail":            tail,
		"chainsawProcess": "txExits",
	})

	defer func() {
		if err != nil {
			wg.Done()
			return
		}

		if err = c.storage.SaveDepositExitPoll(head); err != nil {
			log.WithError(logFields, err).Error("failed to persist deposit exit poll")
		}

		wg.Done()
	}()

	if tail > head {
		logFields.Warn("head is behind last block, implies bug")
		return
	}

	events, _, err := c.client.StartedDepositExitFilter(tail, head)
	if err != nil {
		log.WithError(logFields, err).Error("failed to filter deposit exits")
	}
	if len(events) == 0 {
		logFields.Info("no deposit exits found")
		return
	}

	logFields.WithFields(logrus.Fields{
		"exitCount": len(events),
	}).Info("found deposit exits, checking")

	//TODO: check for race here between incoming transactions, potentially by requerying older blocks
	for _, event := range events {
		nonce := event.Nonce
		evFields := logrus.Fields{
			"nonce":  nonce.Text(10),
			"amount": event.Amount.Text(10),
			"owner":  event.Owner.Hex(),
		}
		challengingTx, err := c.storage.FindDoubleSpendingDeposit(nonce)
		if err != nil {
			log.WithError(logFields, err).WithFields(evFields).Error("failed to query double spends")
			return
		}
		if challengingTx == nil {
			logFields.WithFields(evFields).Info("transaction is not double spent")
			continue
		}
		logFields.WithFields(evFields).Info("found double spend, generating proof")
		exitingBody := chain.ZeroBody()
		exitingBody.Input0.DepositNonce = nonce
		exitingTx := &chain.ConfirmedTransaction{
			Transaction: &chain.Transaction{
				Body: exitingBody,
			},
		}

		txsInChallengeBlock, err := c.storage.FindTransactionsByBlockNum(challengingTx.Transaction.Body.BlockNumber)
		if err != nil {
			log.WithError(logFields, err).WithFields(evFields).Error("failed to query transactions in block")
			return
		}

		// note: proof is expected to be nil for single-transaction proofs
		proof := genTxMerkleProof(txsInChallengeBlock, challengingTx.Transaction.Body.TransactionIndex)
		_, err = c.client.Challenge(exitingTx, 0, nonce, challengingTx, proof)
		if err != nil {
			log.WithError(logFields, err).WithFields(evFields).Error("failed to broadcast exit challenge")
		}
	}
}

func genTxMerkleProof(txs []chain.ConfirmedTransaction, txIdx uint32) []byte {
	var hashes []util.Hash
	for _, tx := range txs {
		hashes = append(hashes, tx.Transaction.RLPHash(util.Sha256))
	}
	_, proof := merkle.RootAndProof(hashes, int64(txIdx))
	return proof
}
