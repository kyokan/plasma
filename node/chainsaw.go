package node

import (
	"time"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/log"
	"github.com/sirupsen/logrus"
	"github.com/kyokan/plasma/chain"
	"math/big"
	"github.com/kyokan/plasma/db"
)

var logger = log.ForSubsystem("Chainsaw")

type Chainsaw struct {
	lastBlock uint64
	quit      chan bool
	client    eth.Client
	mPool     *Mempool
	storage   db.PlasmaStorage
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
	lastBlock, err := c.storage.LastDepositEventIdx()
	if err != nil {
		return err
	}
	c.lastBlock = lastBlock

	go func() {
		logger.Info("chainsaw started")
		tick := time.NewTicker(5 * time.Second)

		for {
			select {
			case <-tick.C:
				c.poll()
			case <-c.quit:
				return
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
		return
	}

	logFields := logger.WithFields(logrus.Fields{
		"tail": c.lastBlock,
		"head": head,
	})
	logFields.Info("processing blocks")
	if err := c.processDeposits(logFields, head); err != nil {
		return
	}
	if err := c.storage.SaveDepositEventIdx(head); err != nil {
		logFields.WithFields(logrus.Fields{"err": err}).Error("failed to persist deposit index")
	}
	c.lastBlock = head + 1
}

func (c *Chainsaw) processDeposits(logFields *logrus.Entry, head uint64) error {
	if c.lastBlock >= head {
		logFields.Warn("head is behind last seen block, implies bug")
		return nil
	}

	events, _, err := c.client.DepositFilter(c.lastBlock, head)
	if err != nil {
		logFields.WithFields(logrus.Fields{"err": err}).Error("failed to filter deposits")
		return err
	}
	if len(events) == 0 {
		logFields.Info("no deposits found")
		return nil
	}

	logFields.WithFields(logrus.Fields{"depositCount": len(events)}).Info("found deposits, adding to mempool")
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
		}
	}
	logFields.WithFields(logrus.Fields{"depositCount": len(events)}).Info("added deposits to mempool")
	return nil
}
