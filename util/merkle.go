package util

import (
	"bufio"
	"bytes"
	"math"

	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/keybase/go-codec/codec"
)

type MerkleTree struct {
	Root MerkleNode
}

type MerkleNode struct {
	Right *MerkleNode
	Left  *MerkleNode
	Hash  Hash
}

func (n MerkleNode) ToCbor() ([]byte, error) {
	buf := new(bytes.Buffer)
	bw := bufio.NewWriter(buf)
	hdl := new(codec.CborHandle)
	enc := codec.NewEncoder(bw, hdl)
	err := enc.Encode(n)

	if err != nil {
		return nil, err
	}

	bw.Flush()

	return buf.Bytes(), nil
}

func TreeFromRLPItems(items []RLPHashable) MerkleTree {
	if len(items) == 0 {
		return emptyTree()
	}

	var level []MerkleNode
	level = make([]MerkleNode, len(items))

	for i, item := range items {
		level[i] = MerkleNode{Hash: item.RLPHash()}
	}

	return treeFromLevel16(level)
}

func TreeFromItems(items []Hashable) MerkleTree {
	if len(items) == 0 {
		return emptyTree()
	}

	var level []MerkleNode
	level = make([]MerkleNode, len(items))

	for i, item := range items {
		level[i] = MerkleNode{Hash: item.Hash()}
	}

	return treeFromLevel(level)
}

func emptyTree() MerkleTree {
	empty := sha3.Sum256(make([]byte, 0))
	return MerkleTree{Root: MerkleNode{Hash: empty[:]}}
}

func treeFromLevel16(level []MerkleNode) MerkleTree {
	if float64(len(level)) > math.Pow(2, 15) {
		// TODO: throw err instead.
		panic("Level must fit within a tree of depth 16!")
	}

	emptyHash := DoHash(make([]byte, 32))

	// Always hash 16 levels.
	for i := 0; i < 15; i++ {
		var nextLevel []MerkleNode

		for i := 0; i < len(level); i += 2 {
			left := &level[i]
			var right *MerkleNode

			// Empty so add an empty node.
			if i+1 > len(level)-1 {
				right = &MerkleNode{
					Hash: emptyHash,
				}
			} else {
				right = &level[i+1]
			}

			nextLevel = append(nextLevel, MerkleNode{
				Left:  left,
				Right: right,
				Hash:  DoHash(append(left.Hash, right.Hash...)),
			})
		}

		level = nextLevel
	}

	return MerkleTree{Root: level[0]}
}

func treeFromLevel(level []MerkleNode) MerkleTree {
	if len(level) == 1 {
		return MerkleTree{Root: level[0]}
	}

	if len(level) == 2 {
		root := MerkleNode{Left: &level[0], Right: &level[1], Hash: hashChildren(&level[0], &level[1])}
		return MerkleTree{Root: root}
	}

	for len(level) > 2 {
		var left *MerkleNode
		var right *MerkleNode

		var nextLevel []MerkleNode

		for _, item := range level {
			if left == nil {
				left = &item
				continue
			}

			if right == nil {
				right = &item
			}

			nextLevel = append(nextLevel, MerkleNode{Right: left, Left: right, Hash: hashChildren(left, right)})
			left = nil
			right = nil
		}

		if left != nil {
			nextLevel = append(nextLevel, *left)
		}

		level = nextLevel
	}

	top := level[0]
	root := MerkleNode{
		Left:  top.Left,
		Right: top.Right,
		Hash:  hashChildren(top.Left, top.Right),
	}

	return MerkleTree{Root: root}
}

func hashChildren(left *MerkleNode, right *MerkleNode) Hash {
	var digest [32]byte
	sha := sha3.New256()
	sha.Write(left.Hash)
	sha.Write(right.Hash)
	return sha.Sum(digest[:])
}
