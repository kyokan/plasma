package merkle

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyokan/plasma/util"
	"math"
	"sync"
)

const (
	maxDepth = 17
)

type DualHashable interface {
	util.Hashable
	util.RLPHashable
	SetIndex(uint32)
}

type Hasher func([]byte) util.Hash

var precomputedHashes [maxDepth]util.Hash
var once sync.Once

// Initialized instead of hard-coded values as the hash function may change
func precomputeHashes(hasher Hasher, sparse bool) {
	emptyBuf := make([]byte, 32)
	var index int
	precomputedHashes   [maxDepth - 1] = hasher(emptyBuf)

	if sparse {
		for index = maxDepth - 2; index >= 0; index-- {
			precomputedHashes[index] = precomputedHashes[maxDepth - 1]
		}
		return
	}

	for index = maxDepth - 2; index >= 0; index-- {
		precomputedHashes[index]    = hasher(append(precomputedHashes[index+1], precomputedHashes[index+1]...))
	}
}

type MerkleQueue interface {
	Enqueue(transaction []DualHashable) error
	GetRootRLPHash() (util.Hash, error)
	GetRootHash() (util.Hash, error)
	GetRLPProof() ([]util.Hash, error)
	Reset()
	GetNumberOfLeafes() uint32
}

type queueElement struct {
	depth   int32
	hash    util.Hash
	rlpHash util.Hash
}

func createQueueElement(depth int32, hash, rlpHash util.Hash) *queueElement {
	result := queueElement{
		depth:   depth,
		hash:    hash,
		rlpHash: rlpHash,
	}
	return &result
}

type proofState struct {
	indices   []int32
	hashes    []util.Hash
	rlpHashes []util.Hash
}

func createProofState(index, depth int32) *proofState {
	result := &proofState{
	}
	levelIdx := index
	result.indices   = make([]int32, depth + 1)
	for i := int32(depth); i > 0; i-- {
		if levelIdx % 2 == 0 { // If the index is even, the right sibling is part of the proof
			result.indices[i] = levelIdx + 1
		} else { // If the index is odd, the left sibling is part of the proof
			result.indices[i] = levelIdx - 1
		}
		levelIdx = levelIdx / 2
	}
	result.hashes    = make([]util.Hash, depth+1)
	result.rlpHashes = make([]util.Hash, depth+1)
	return result
}

func (state *proofState) store(element *queueElement) {
	if state == nil || state.indices[element.depth] < 0{
		return
	}
	if state.indices[element.depth] == 0 {
		state.hashes[element.depth] = element.hash
		state.rlpHashes[element.depth] = element.rlpHash
	}
	state.indices[element.depth]--
}

type merkleQueue struct {
	sync.Mutex

	current     []*queueElement
	depth       int32
	context     context.Context
	cancel      context.CancelFunc
	leaves      uint32
	leafIndex   uint32
	maxLeaves   float64
	proofIndex  int32

	proofState *proofState

	queue chan DualHashable

	hasher Hasher
}

func createMerkleQueue(hasher Hasher, depth, index int32, sparse bool) (*merkleQueue, error) {
	once.Do(func() {
		precomputeHashes(hasher, sparse)
	})
	if depth < 2 {
		return nil, errors.New("Depth has to be at least 2")
	}
	if depth > maxDepth {
		return nil, errors.New(fmt.Sprintf("Depth can be at most %d", maxDepth))
	}
	result := merkleQueue{
		current:    make([]*queueElement, depth),
		depth:      depth - 1, // zero based
		leaves:     uint32(0),
		leafIndex:  uint32(0),
		maxLeaves:  math.Pow(2, float64(depth-1)),
		proofIndex: index,
		hasher:     hasher,
		queue:      make(chan DualHashable),
	}
	result.context, result.cancel = context.WithCancel(context.Background())

	if index >= 0 {
		result.proofState = createProofState(result.proofIndex, result.depth)
	}
	go result.listen()
	return &result, nil
}

func NewMerkleQueue(depth int32, sparse bool) (MerkleQueue, error) {
	return createMerkleQueue(util.DoHash, depth, -1, sparse)
}

func (merkle *merkleQueue) processElement(right *queueElement) {
	var left *queueElement
	for right.depth >= 0 {
		left = merkle.current[right.depth] // sibling node
		if left == nil { // if current for this depth is empty we store & stop
			merkle.current[right.depth] = right
			break
		}
		// Otherwise we keep going up the tree
		// Set the current for this depth to empty
		merkle.current[right.depth] = nil
		// compute the element for the next level
		hash := merkle.hasher(append(left.hash, right.hash...))
		rlpHash := merkle.hasher(append(left.rlpHash, right.rlpHash...))
		right = createQueueElement(right.depth-1, hash, rlpHash)
		merkle.proofState.store(right)
	}
}

func (merkle *merkleQueue) getPrecomputedHash(elementDepth int32) util.Hash {
	diff := maxDepth - 1 - merkle.depth
	return precomputedHashes[elementDepth + diff]
}

func (merkle *merkleQueue) computeRootHashes() error {
	merkle.cancel()
	merkle.Lock()
	defer merkle.Unlock()

	if merkle.leaves == 0 {
		merkle.current[0] = createQueueElement(merkle.depth,
			merkle.getPrecomputedHash(0),
			merkle.getPrecomputedHash(0))
		return nil
	}

	var depth = merkle.depth
	for ; depth >= 0; depth-- {
		if merkle.current[depth] != nil {
			break
		}
	}
	if depth <= 0 {
		return nil
	}

	right := createQueueElement(depth,
		merkle.getPrecomputedHash(depth),
		merkle.getPrecomputedHash(depth))
	merkle.proofState.store(right)

	var left *queueElement
	for ;depth > 0; depth-- {
		left = merkle.current[depth] // this should non-null for the first iteration
		/*
		If there's no sibling, use a precomputed one.
		Also swap left and right, as precomputed values are on the right-hand side.
		 */
		if left == nil {
			left  = right
			right = createQueueElement(depth,
				merkle.getPrecomputedHash(depth),
				merkle.getPrecomputedHash(depth))
			merkle.proofState.store(right)
		}
		merkle.current[depth] = nil

		right = createQueueElement(depth-1,
			merkle.hasher(append(left.hash, right.hash...)),
			merkle.hasher(append(left.rlpHash, right.rlpHash...)))
		merkle.proofState.store(right)
	}
	merkle.current[0] = right

	return nil
}

func (merkle *merkleQueue) listen() {
	merkle.Lock()
	defer merkle.Unlock()
	for {
		select {
		case tx := <-merkle.queue:
			tx.SetIndex(merkle.leafIndex)
			merkle.leafIndex++
			newElement := createQueueElement(merkle.depth, tx.Hash(), tx.RLPHash())
			merkle.proofState.store(newElement)
			merkle.processElement(newElement)

		case <-merkle.context.Done():
			return
		}
	}
}

func (merkle *merkleQueue) Reset() {
	for i := int32(0); i < merkle.depth; i++ {
		merkle.current[i] = nil
	}
	merkle.leaves = uint32(0)
	merkle.leafIndex = uint32(0)
	merkle.context, merkle.cancel = context.WithCancel(context.Background())

	if merkle.proofIndex >= 0 {
		merkle.proofState = createProofState(merkle.proofIndex, merkle.depth)
	}
	go merkle.listen()
}

func (merkle *merkleQueue) Enqueue(transactions []DualHashable) error {
	if float64(merkle.leaves + uint32(len(transactions))) > merkle.maxLeaves {
		return errors.New("Max number of leaves reached")
	}
	for i := 0; i < len(transactions); i++ {
		merkle.leaves++
		merkle.queue <- transactions[i]
	}

	return nil
}

func (merkle *merkleQueue) GetRootHash() (util.Hash, error) {
	err := merkle.computeRootHashes()
	if err != nil {
		return nil, err
	}
	return merkle.current[0].hash, nil
}

func (merkle *merkleQueue) GetRootRLPHash() (util.Hash, error) {
	err := merkle.computeRootHashes()
	if err != nil {
		return nil, err
	}
	return merkle.current[0].rlpHash, nil
}

func (merkle *merkleQueue) GetRLPProof() ([]util.Hash, error) {
	err := merkle.computeRootHashes()
	if err != nil {
		return nil, err
	}
	length := len(merkle.proofState.rlpHashes) - 1 // root is not included in proof
	result := make([]util.Hash, length)
	for i := 0; i < length; i++ {
		result[i] = merkle.proofState.rlpHashes[length - i]
	}
	return result, nil
}

func (merkle *merkleQueue) GetNumberOfLeafes() uint32 {
	return merkle.leaves
}

func doGetProof(transactions []DualHashable, hasher Hasher, depth, index int32) ([]util.Hash, error) {
	queue, err := createMerkleQueue(hasher, depth, index, false)
	if err != nil {
		return nil, err
	}
	err = queue.Enqueue(transactions)
	if err != nil {
		return nil, err
	}
	return queue.GetRLPProof()
}

func GetProof(transactions []DualHashable, depth, index int32) ([]util.Hash, error) {
	return doGetProof(transactions, util.DoHash, depth, index)
}

