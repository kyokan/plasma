package rpc

import (
	"github.com/kyokan/plasma/rpc/pb"
	"math/big"
	"github.com/kyokan/plasma/chain"
	"github.com/ethereum/go-ethereum/common"
	"strings"
	"fmt"
)

func SerializeBig(in *big.Int) (*pb.BigInt) {
	return &pb.BigInt{
		Hex: fmt.Sprintf("0x%s", strings.ToLower(in.Text(16))),
	}
}

func DeserializeBig(in *pb.BigInt) (*big.Int) {
	b, _ := new(big.Int).SetString(in.Hex, 16)
	return b
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
		Input0: SerializeInput(tx.Input0),
		Sig0: tx.Sig0[:],
		Input1: SerializeInput(tx.Input1),
		Sig1: tx.Sig1[:],
		Output0: SerializeOutput(tx.Output0),
		Output1: SerializeOutput(tx.Output1),
		Fee: SerializeBig(tx.Fee),
		BlockNum: SerializeBig(tx.BlkNum),
		TxIdx: SerializeBig(tx.TxIdx),
	}
}

func DeserializeTx(tx *pb.Transaction) (*chain.Transaction) {
	result := chain.ZeroTransaction()
	if tx != nil {
		result.Input0  = DeserializeInput(tx.Input0)
		copy(result.Sig0[:], tx.Sig0)
		result.Input1  = DeserializeInput(tx.Input1)
		copy(result.Sig1[:], tx.Sig1)
		result.Output0 = DeserializeOutput(tx.Output0)
		result.Output1 = DeserializeOutput(tx.Output1)
		result.Fee     = DeserializeBig(tx.Fee)
		result.BlkNum  = DeserializeBig(tx.BlockNum)
		result.TxIdx   = DeserializeBig(tx.TxIdx)
		copy(result.Sig0[:], tx.Sig0)
	}
	return result
}

func SerializeInput(in *chain.Input) (*pb.Input) {
	return &pb.Input{
		BlockNum: SerializeBig(in.BlkNum),
		TxIdx:    SerializeBig(in.TxIdx),
		OutIdx:   SerializeBig(in.OutIdx),
	}
}

func DeserializeInput(in *pb.Input) (*chain.Input) {
	return &chain.Input{
		BlkNum: DeserializeBig(in.BlockNum),
		TxIdx:  DeserializeBig(in.TxIdx),
		OutIdx: DeserializeBig(in.OutIdx),
	}
}

func SerializeOutput(out *chain.Output) (*pb.Output) {
	return &pb.Output{
		NewOwner:     out.Owner.Bytes(),
		Amount:       SerializeBig(out.Denom),
		DepositNonce: SerializeBig(out.DepositNonce),
	}
}

func DeserializeOutput(out *pb.Output) (*chain.Output) {
	return &chain.Output{
		Owner:        common.BytesToAddress(out.NewOwner),
		Denom:        DeserializeBig(out.Amount),
		DepositNonce: DeserializeBig(out.DepositNonce),
	}
}

func DeserializeBlock(block *pb.Block) (*chain.Block) {
	return &chain.Block{
		Header: &chain.BlockHeader{
			MerkleRoot:    block.Header.MerkleRoot,
			RLPMerkleRoot: block.Header.RlpMerkleRoot,
			PrevHash:      block.Header.PrevHash,
			Number:        block.Header.Number,
		},
		BlockHash: block.Hash,
	}
}

func De0x(in string) string {
	return strings.Replace(in, "0x", "", 1)
}
