package chain

import (
	"bytes"
	"encoding/binary"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/kyokan/plasma/util"
)

// JSON tags needed for test fixtures
type BlockHeader struct {
	MerkleRoot    util.Hash `json:"MerkleRoot"`
	RLPMerkleRoot util.Hash `json:"RLPMerkleRoot"`
	PrevHash      util.Hash `json:"PrevHash"`
	Number        uint64    `json:"Number"`
}

// JSON tags needed for test fixtures
type Block struct {
	Header    *BlockHeader `json:"Header"`
	BlockHash util.Hash    `json:"BlockHash"`
}

func (head BlockHeader) Hash() util.Hash {
	buf := new(bytes.Buffer)
	buf.Write(head.MerkleRoot)
	buf.Write(head.PrevHash)
	binary.Write(buf, binary.BigEndian, head.Number)
	digest := sha3.Sum256(buf.Bytes())
	return digest[:]
}
