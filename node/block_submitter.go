package node

import (
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"sync"
	"sync/atomic"
	log2 "github.com/kyokan/plasma/log"
	"github.com/sirupsen/logrus"
	"github.com/kyokan/plasma/util"
	"github.com/syndtr/goleveldb/leveldb"
)

var bsLogger = log2.ForSubsystem("BlockSubmitter")

type BlockSubmitter struct {
	submissions  []db.BlockResult
	awakeDequeue chan bool
	client       eth.Client
	mtx          sync.Mutex
	isBusy       uint32
	ps           db.PlasmaStorage
}

func NewBlockSubmitter(client eth.Client, ps db.PlasmaStorage) *BlockSubmitter {
	res := &BlockSubmitter{
		submissions:  make([]db.BlockResult, 0),
		awakeDequeue: make(chan bool),
		client:       client,
		ps:           ps,
	}

	go func() {
		for {
			<-res.awakeDequeue
			go res.dequeue()
		}
	}()

	return res
}

func (s *BlockSubmitter) Start() error {
	lastSubmitted, err := s.ps.LastSubmittedBlock()
	if err == leveldb.ErrNotFound {
		return nil
	}
	latest, err := s.ps.LatestBlock()
	if latest == nil {
		return nil
	}
	if lastSubmitted == latest.Header.Number {
		return nil
	}

	toEnqueue := make([]db.BlockResult, 0)
	for i := lastSubmitted + 1; i <= latest.Header.Number; i++ {
		block, err := s.ps.BlockAtHeight(i)
		if err != nil {
			return err
		}
		meta, err := s.ps.BlockMetaAtHeight(i)
		if err != nil {
			return err
		}

		toEnqueue = append(toEnqueue, db.BlockResult{
			MerkleRoot:         block.Header.MerkleRoot,
			NumberTransactions: meta.TransactionCount,
			BlockFees:          meta.Fees,
			BlockNumber:        util.Uint642Big(i),
		})
	}

	for _, res := range toEnqueue {
		s.Enqueue(res)
	}

	return nil
}

func (s *BlockSubmitter) Stop() error {
	return nil
}

func (s *BlockSubmitter) Enqueue(res db.BlockResult) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.submissions = append(s.submissions, res)
	s.awakeDequeue <- true
	bsLogger.WithFields(logrus.Fields{
		"blockNumber": util.Big2Uint64(res.BlockNumber),
	}).Info("enqueued block for submission")
}

func (s *BlockSubmitter) dequeue() {
	if !atomic.CompareAndSwapUint32(&s.isBusy, 0, 1) {
		return
	}
	defer atomic.StoreUint32(&s.isBusy, 0)
	if len(s.submissions) == 0 {
		return
	}

	s.mtx.Lock()
	latest := s.submissions[0]
	s.submissions = s.submissions[1:]
	s.mtx.Unlock()

	err := s.client.SubmitBlock(latest.MerkleRoot, latest.NumberTransactions, latest.BlockFees, latest.BlockNumber)
	if err != nil {
		bsLogger.WithFields(logrus.Fields{
			"blockNumber": util.Big2Uint64(latest.BlockNumber),
			"err":         err,
		}).Error("failed to submit block!")
		s.mtx.Lock()
		s.submissions = append([]db.BlockResult{latest}, s.submissions...)
		s.mtx.Unlock()
		return
	}

	bnUint := util.Big2Uint64(latest.BlockNumber)
	bsLogger.WithFields(logrus.Fields{
		"blockNumber": bnUint,
	}).Info("successfully submitted block")

	err = s.ps.SaveLastSubmittedBlock(bnUint)
	if err != nil {
		bsLogger.WithFields(logrus.Fields{
			"blockNumber": bnUint,
			"err":         err,
		}).Error("failed to persist last submitted block number")
	}

	s.awakeDequeue <- true
}
