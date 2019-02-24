package chain

import (
	"bytes"
	"encoding/binary"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/kyokan/plasma/util"
	"math/big"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/pkg/rpc/pb"
	"github.com/kyokan/plasma/pkg/rpc"
)

type BlockHeader struct {
	MerkleRoot    util.Hash `json:"merkleRoot"`
	RLPMerkleRoot util.Hash `json:"rlpMerkleRoot"`
	PrevHash      util.Hash `json:"prevHash"`
	Number        uint64    `json:"number"`
}

type Block struct {
	Header    *BlockHeader `json:"header"`
	BlockHash util.Hash    `json:"blockHash"`
}

type BlockMetadata struct {
	CreatedAt        uint64   `json:"createdAt"`
	TransactionCount uint32   `json:"transactionCount"`
	Fees             *big.Int `json:"fees"`
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

func BlockFromProto(protoBlock *pb.Block) *Block {
	return &Block{
		Header: &BlockHeader{
			MerkleRoot:    protoBlock.Header.MerkleRoot,
			RLPMerkleRoot: protoBlock.Header.RlpMerkleRoot,
			PrevHash:      protoBlock.Header.PrevHash,
			Number:        protoBlock.Header.Number,
		},
		BlockHash: protoBlock.Hash,
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
		CreatedAt:        m.CreatedAt,
		TransactionCount: m.TransactionCount,
		Fees:             rpc.SerializeBig(m.Fees),
	}
}

func BlockMetadataFromProto(m *pb.BlockMeta) *BlockMetadata {
	return &BlockMetadata{
		CreatedAt:        m.CreatedAt,
		TransactionCount: m.TransactionCount,
		Fees:             rpc.DeserializeBig(m.Fees),
	}
}

func (head *BlockHeader) Hash() util.Hash {
	buf := new(bytes.Buffer)
	buf.Write(head.MerkleRoot)
	buf.Write(head.PrevHash)
	binary.Write(buf, binary.BigEndian, head.Number)
	digest := sha3.Sum256(buf.Bytes())
	return digest[:]
}
