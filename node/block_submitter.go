package node

import (
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"sync"
		"sync/atomic"
	log2 "github.com/kyokan/plasma/log"
	"github.com/sirupsen/logrus"
	"github.com/kyokan/plasma/util"
)

var bsLogger = log2.ForSubsystem("BlockSubmitter")

type BlockSubmitter struct {
	submissions  []db.BlockResult
	awakeDequeue chan bool
	client       eth.Client
	mtx          sync.Mutex
	isBusy       uint32
}

func NewBlockSubmitter(client eth.Client) *BlockSubmitter {
	res := &BlockSubmitter{
		submissions:  make([]db.BlockResult, 0),
		awakeDequeue: make(chan bool),
		client:       client,
	}

	go func() {
		for {
			<-res.awakeDequeue
			go res.dequeue()
		}
	}()

	return res
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
			"err": err,
		}).Error("failed to submit block!")
		s.mtx.Lock()
		s.submissions = append([]db.BlockResult{latest}, s.submissions...)
		s.mtx.Unlock()
		return
	}

	bsLogger.WithFields(logrus.Fields{
		"blockNumber": util.Big2Uint64(latest.BlockNumber),
	}).Info("successfully submitted block")
	s.awakeDequeue <- true
}
