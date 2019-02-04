package node

import (
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"sync"
	"log"
	"sync/atomic"
)

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
	log.Printf("enqueued submission for block %d", res.BlockNumber.Uint64())
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
		log.Println("error submitting block", err)
		s.mtx.Lock()
		s.submissions = append([]db.BlockResult{latest}, s.submissions...)
		s.mtx.Unlock()
		return
	}

	log.Printf("successfully submitted block %d", latest.BlockNumber.Uint64())

	s.awakeDequeue <- true
}
