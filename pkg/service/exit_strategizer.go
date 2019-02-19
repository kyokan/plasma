package service

import (
	"github.com/kyokan/plasma/pkg/eth"
	"github.com/kyokan/plasma/pkg/log"
	"time"
	"github.com/sirupsen/logrus"
	"sync/atomic"
	"github.com/kyokan/plasma/pkg/db"
	"github.com/kyokan/plasma/pkg/merkle"
	"github.com/kyokan/plasma/util"
)

var exitStratLogger = log.ForSubsystem("ExitStrategizer")

type ExitStrategizer struct {
	ethClient   eth.Client
	storage     db.Storage
	mainBreaker CircuitBreaker

	lastRootResponse time.Time
	corruptBlocks    uint64
	exiting          uint32

	blockCorrupted   chan bool
	rootResponsive   chan bool
	rootUnresponsive chan bool
	quit             chan bool
}

func NewExitStrategizer(ethClient eth.Client, storage db.Storage, mainBreaker CircuitBreaker) *ExitStrategizer {
	return &ExitStrategizer{
		ethClient:        ethClient,
		storage:          storage,
		mainBreaker:      mainBreaker,
		blockCorrupted:   make(chan bool),
		rootResponsive:   make(chan bool),
		rootUnresponsive: make(chan bool),
		quit:             make(chan bool),
	}
}

func (e *ExitStrategizer) Start() error {
	go func() {
		for {
			select {
			case <-e.blockCorrupted:
				e.corruptBlocks++
				e.maybeExit()
			case <-e.rootResponsive:
				e.lastRootResponse = time.Now()
				e.maybeExit()
			case <-e.rootUnresponsive:
				e.maybeExit()
			case <-e.quit:
				return
			}
		}
	}()

	return nil
}

func (e *ExitStrategizer) Stop() error {
	e.quit <- true
	return nil
}

func (e *ExitStrategizer) RootUnresponsive() {
	exitStratLogger.Warn("root is unresponsive")
	e.rootUnresponsive <- true
}

func (e *ExitStrategizer) RootResponsive() {
	e.rootResponsive <- true
}

func (e *ExitStrategizer) BlockCorrupted() {
	exitStratLogger.Warn("received corrupted block")
	e.blockCorrupted <- true
}

func (e *ExitStrategizer) maybeExit() {
	if e.corruptBlocks > 0 {
		exitStratLogger.Warn("root node provided corrupt block, performing mass exit")
		e.doExit()
		return
	}

	if time.Since(e.lastRootResponse) > 30*time.Second {
		exitStratLogger.Warn("have not seen root node for past 10 minutes, performing mass exit")
		e.doExit()
		return
	}
}

func (e *ExitStrategizer) doExit() {
	if atomic.CompareAndSwapUint32(&e.exiting, 0, 1) {
		exitStratLogger.Info("already exiting")
	}

	e.mainBreaker.Trip()
	addr := e.ethClient.UserAddress()
	exitStratLogger.WithFields(logrus.Fields{
		"address": addr.Hex(),
	}).Info("exiting all UTXOs")

	utxos, err := e.storage.UTXOs(addr)
	if err != nil {
		log.WithError(exitStratLogger, err).Error("failed to fetch UTXOs for user!")
		return
	}

	for _, utxo := range utxos {
		txs, err := e.storage.FindTransactionsByBlockNum(utxo.Transaction.Body.BlockNumber)
		if err != nil {
			log.WithError(exitStratLogger, err).Error("couldn't find transactions inside block while exiting transaction")
			continue
		}

		leaves := make([]util.Hash, len(txs), len(txs))
		for i, tx := range txs {
			leaves[i] = tx.Transaction.RLPHash(util.Sha256)
		}
		_, proof := merkle.RootAndProof(leaves, int64(utxo.Transaction.Body.TransactionIndex))
		_, err = e.ethClient.Exit(&utxo, utxo.Transaction.Body.OutputIndexFor(&addr), proof)
		if err != nil {
			log.WithError(exitStratLogger, err).Error("failed to perform exit")
			continue
		}
		exitStratLogger.Info("successfully exited output")
	}

	logger.Warn("mass exit complete. you can shut down your node now.")
}
