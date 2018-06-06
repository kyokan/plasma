package chain

import (
	"bytes"
	"encoding/binary"

	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/kyokan/plasma/util"
)

type BlockHeader struct {
	MerkleRoot    util.Hash
	RLPMerkleRoot util.Hash
	PrevHash      util.Hash
	Number        uint64
}

type Block struct {
	Header    *BlockHeader
	BlockHash util.Hash
}

func (head BlockHeader) Hash() util.Hash {
	buf := new(bytes.Buffer)
	buf.Write(head.MerkleRoot)
	buf.Write(head.PrevHash)
	binary.Write(buf, binary.BigEndian, head.Number)
	digest := sha3.Sum256(buf.Bytes())
	return digest[:]
}
