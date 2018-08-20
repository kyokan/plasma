package rpc

import (
	"github.com/kyokan/plasma/rpc/pb"
	"math/big"
	"github.com/kyokan/plasma/chain"
	"github.com/ethereum/go-ethereum/common"
)

func DeserializeBig(in *pb.BigInt) (*big.Int) {
	return new(big.Int).SetBytes(in.Values)
}

func DeserializeUintBig(in *pb.BigInt) uint {
	return uint(DeserializeBig(in).Uint64())
}

func SerializeBig(in *big.Int) (*pb.BigInt) {
	return &pb.BigInt{
		Values: in.Bytes(),
	}
}

func SerializeUintBig(in uint64) (*pb.BigInt) {
	bi := new(big.Int).SetUint64(in)

	return &pb.BigInt{
		Values: bi.Bytes(),
	}
}

func SerializeTxs(txs []chain.Transaction) ([]*pb.Transaction) {
	out := make([]*pb.Transaction, len(txs), len(txs))
	for i, tx := range txs {
		out[i] = SerializeTx(&tx)
	}
	return out
}

func DeserializeTxs(txs []*pb.Transaction) ([]chain.Transaction) {
	out := make([]chain.Transaction, len(txs), len(txs))
	for i, tx := range txs {
		out[i] = *DeserializeTx(tx)
	}
	return out
}

func SerializeTx(tx *chain.Transaction) (*pb.Transaction) {
	return &pb.Transaction{
		Input0:   SerializeInput(tx.Input0),
		Sig0:     tx.Sig0,
		Input1:   SerializeInput(tx.Input1),
		Sig1:     tx.Sig1,
		Output0:  SerializeOutput(tx.Output0),
		Output1:  SerializeOutput(tx.Output1),
		Fee:      SerializeBig(tx.Fee),
		BlockNum: SerializeUintBig(tx.BlkNum),
		TxIdx:    tx.TxIdx,
	}
}

func DeserializeTx(tx *pb.Transaction) (*chain.Transaction) {
	return &chain.Transaction{
		Input0: DeserializeInput(tx.Input0),
		Sig0: tx.Sig0,
		Input1: DeserializeInput(tx.Input1),
		Sig1:tx.Sig1,
		Output0: DeserializeOutput(tx.Output0),
		Output1: DeserializeOutput(tx.Output1),
		Fee: DeserializeBig(tx.Fee),
		BlkNum: uint64(DeserializeUintBig(tx.BlockNum)),
		TxIdx: tx.TxIdx,
	}
}

func SerializeInput(in *chain.Input) (*pb.Input) {
	return &pb.Input{
		BlockNum: SerializeUintBig(in.BlkNum),
		TxIdx:    in.TxIdx,
		OutIdx:   uint32(in.OutIdx),
	}
}

func DeserializeInput(in *pb.Input) (*chain.Input) {
	return &chain.Input{
		BlkNum: uint64(DeserializeUintBig(in.BlockNum)),
		TxIdx: in.TxIdx,
		OutIdx: uint8(in.TxIdx),
	}
}

func SerializeOutput(out *chain.Output) (*pb.Output) {
	return &pb.Output{
		NewOwner: out.NewOwner.Bytes(),
		Amount:   SerializeBig(out.Amount),
	}
}

func DeserializeOutput(out *pb.Output) (*chain.Output) {
	return &chain.Output{
		NewOwner: common.BytesToAddress(out.NewOwner),
		Amount: DeserializeBig(out.Amount),
	}
}

func DeserializeBlock(block *pb.Block) (*chain.Block) {
	return &chain.Block{
		Header: &chain.BlockHeader{
			MerkleRoot:block.Header.MerkleRoot,
			RLPMerkleRoot:block.Header.RlpMerkleRoot,
			PrevHash:block.Header.PrevHash,
			Number: uint64(DeserializeUintBig(block.Header.Number)),
		},
		BlockHash: block.Hash,
	}
}