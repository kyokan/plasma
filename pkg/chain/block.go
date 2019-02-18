package chain

import (
	"bytes"
	"encoding/binary"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/kyokan/plasma/util"
	"math/big"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/pkg/rpc/pb"
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

func (block *Block) Proto() *pb.Block {
	return &pb.Block{
		Header: &pb.BlockHeader{
			MerkleRoot:    block.Header.MerkleRoot,
			RlpMerkleRoot: block.Header.RLPMerkleRoot,
			PrevHash:      block.Header.PrevHash,
			Number:        block.Header.Number,
		},
		Hash: block.BlockHash,
	}
}

func (m *BlockMetadata) RLP() ([]byte, error) {
	rlpVer := &rlpBlockMetadata{
		CreatedAt:        m.CreatedAt,
		TransactionCount: m.TransactionCount,
		Fees:             NewUint256(m.Fees),
	}

	return rlp.EncodeToBytes(rlpVer)
}

func (m *BlockMetadata) FromRLP(data []byte) error {
	var rlpVer rlpBlockMetadata
	err := rlp.DecodeBytes(data, &rlpVer)
	if err != nil {
		return err
	}

	m.CreatedAt = rlpVer.CreatedAt
	m.TransactionCount = rlpVer.TransactionCount
	m.Fees = rlpVer.Fees.ToBig()
	return nil
}

func (m *BlockMetadata) Proto() *pb.BlockMeta {
	return &pb.BlockMeta{
		CreatedAt: m.CreatedAt,
	}
}

func (head BlockHeader) Hash() util.Hash {
	buf := new(bytes.Buffer)
	buf.Write(head.MerkleRoot)
	buf.Write(head.PrevHash)
	binary.Write(buf, binary.BigEndian, head.Number)
	digest := sha3.Sum256(buf.Bytes())
	return digest[:]
}
