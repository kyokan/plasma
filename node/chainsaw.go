package node

import (
	"time"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/log"
	"github.com/sirupsen/logrus"
	"github.com/kyokan/plasma/chain"
	"math/big"
	"github.com/kyokan/plasma/db"
	"sync"
	"github.com/kyokan/plasma/util"
	"github.com/kyokan/plasma/merkle"
	)

var logger = log.ForSubsystem("Chainsaw")

type Chainsaw struct {
	quit    chan bool
	client  eth.Client
	mPool   *Mempool
	storage db.PlasmaStorage
}

func NewChainsaw(client eth.Client, mPool *Mempool, storage db.PlasmaStorage) *Chainsaw {
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
				if duration < 5 * time.Second {
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
	go c.processDeposits(&wg, head)
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

		txsInChallengeBlock, err := c.storage.FindTransactionsByBlockNum(util.Big2Uint64(challengingTx.Transaction.BlkNum))
		if err != nil {
			log.WithError(logFields, err).WithFields(evFields).Error("failed to query transactions in block")
			return
		}
		authSigs, err := c.storage.AuthSigsFor(blkNum, txIdx)
		if err != nil {
			log.WithError(logFields, err).WithFields(evFields).Error("failed to query authSigs")
			return
		}

		// note: proof is expected to be nil for single-transaction proofs
		proof := genTxMerkleProof(txsInChallengeBlock, txIdx)

		_, err = c.client.Challenge(exitingTx, outIdx, big.NewInt(0), challengingTx, proof, authSigs[0])
		if err != nil {
			log.WithError(logFields, err).WithFields(evFields).Error("failed to broadcast exit challenge")
		}
	}
}

func (c *Chainsaw) processDeposits(wg *sync.WaitGroup, head uint64) {
	var err error
	tail, err := c.storage.LastDepositPoll()
	if err != nil {
		wg.Done()
		log.WithError(logger, err).Error("failed to fetch last seen block in transaction exit poller")
		return
	}
	tail += 1

	logFields := logger.WithFields(logrus.Fields{
		"head":            head,
		"tail":            tail,
		"chainsawProcess": "deposits",
	})

	defer func() {
		if err != nil {
			wg.Done()
			return
		}

		if err = c.storage.SaveDepositPoll(head); err != nil {
			log.WithError(logFields, err).Error("failed to persist deposit poll")
		}
		wg.Done()
	}()

	if tail > head {
		logFields.Warn("head is behind last block, implies bug")
		return
	}

	events, _, err := c.client.DepositFilter(tail, head)
	if err != nil {
		logFields.WithFields(logrus.Fields{"err": err}).Error("failed to filter deposits")
		return
	}
	if len(events) == 0 {
		logFields.Info("no deposits found")
		return
	}

	logFields.WithFields(logrus.Fields{
		"depositCount": len(events),
	}).Debug("found deposits, adding to mempool")

	for _, event := range events {
		tx := chain.Transaction{
			Input0: chain.ZeroInput(),
			Input1: chain.ZeroInput(),
			Output0: &chain.Output{
				Owner:        event.Depositor,
				Denom:        event.Amount,
				DepositNonce: event.DepositNonce,
			},
			Output1: chain.ZeroOutput(),
			Fee:     big.NewInt(0),
		}
		confirmed := chain.ConfirmedTransaction{Transaction: tx,}
		inclusion := c.mPool.Append(confirmed)
		if inclusion.Error != nil {
			logFields.WithFields(logrus.Fields{"err": err, "txHash": tx.SignatureHash().Hex()}).
				Error("error while adding deposit to mempool")
			// TODO: resume where we left off
			return
		}
	}
	logFields.WithFields(logrus.Fields{"depositCount": len(events)}).Info("added deposits to mempool")
}

func genTxMerkleProof(txs []chain.ConfirmedTransaction, txIdx uint32) []byte {
	var hashes []util.Hash
	for _, tx := range txs {
		hashes = append(hashes, tx.RLPHash(util.Sha256))
	}
	_, proof := merkle.RootAndProof(hashes, int64(txIdx))
	return proof
}
