package chain

import (
	"bytes"
	"encoding/binary"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/kyokan/plasma/util"
	"math/big"
	"github.com/ethereum/go-ethereum/rlp"
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

type BlockMetadata struct {
	CreatedAt        uint64
	TransactionCount uint32
	Fees             *big.Int
}

type rlpBlockMetadata struct {
	CreatedAt        uint64
	TransactionCount uint32
	Fees             *UInt256
}

func (b *BlockMetadata) RLP() ([]byte, error) {
	rlpVer := &rlpBlockMetadata{
		CreatedAt:        b.CreatedAt,
		TransactionCount: b.TransactionCount,
		Fees:             NewUint256(b.Fees),
	}

	return rlp.EncodeToBytes(rlpVer)
}

func (b *BlockMetadata) FromRLP(data []byte) error {
	var rlpVer rlpBlockMetadata
	err := rlp.DecodeBytes(data, &rlpVer)
	if err != nil {
		return err
	}

	b.CreatedAt = rlpVer.CreatedAt
	b.TransactionCount = rlpVer.TransactionCount
	b.Fees = rlpVer.Fees.ToBig()
	return nil
}

func (head BlockHeader) Hash() util.Hash {
	buf := new(bytes.Buffer)
	buf.Write(head.MerkleRoot)
	buf.Write(head.PrevHash)
	binary.Write(buf, binary.BigEndian, head.Number)
	digest := sha3.Sum256(buf.Bytes())
	return digest[:]
}
