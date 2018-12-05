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
	var result pb.Transaction
	if tx.IsDeposit() {
		result.Value = &pb.Transaction_Deposit{
			Deposit: &pb.Deposit{
				DepositNonce: SerializeBig(tx.DepositNonce),
				Amount:       SerializeBig(tx.Amount),
		}}
		return &result
	}
	result.Value = &pb.Transaction_Utxo{
		Utxo: &pb.UTXO{
		Input0:   SerializeInput(tx.Input0),
		Sig0:     tx.Sig0,
		Input1:   SerializeInput(tx.Input1),
		Sig1:     tx.Sig1,
		Output0:  SerializeOutput(tx.Output0),
		Output1:  SerializeOutput(tx.Output1),
		Fee:      SerializeBig(tx.Fee),
		BlockNum: SerializeBig(tx.BlkNum),
		TxIdx:    SerializeBig(tx.TxIdx), 
	}}
	return &result;
}

func DeserializeTx(tx *pb.Transaction) (*chain.Transaction) {
	if tx == nil {
		return chain.ZeroTransaction()
	}
	var result chain.Transaction
	deposit := tx.GetDeposit()
	if deposit != nil {
		result.Deposit = chain.Deposit{
			DepositNonce: DeserializeBig(deposit.DepositNonce), 
			Amount: DeserializeBig(deposit.Amount),
		}
		return &result
	}
	utxo := tx.GetUtxo()
	if utxo != nil {
		result.Input0  = DeserializeInput(utxo.Input0)
		result.Sig0    = utxo.Sig0
		result.Input1  = DeserializeInput(utxo.Input1)
		result.Sig1    = utxo.Sig1
		result.Output0 = DeserializeOutput(utxo.Output0)
		result.Output1 = DeserializeOutput(utxo.Output1)
		result.Fee     = DeserializeBig(utxo.Fee)
		result.BlkNum  = DeserializeBig(utxo.BlockNum)
		result.TxIdx   = DeserializeBig(utxo.TxIdx)
		return &result
	}
	return chain.ZeroTransaction()
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
		NewOwner:     out.NewOwner.Bytes(),
		Amount:       SerializeBig(out.Denom),
		DepositNonce: SerializeBig(out.DepositNonce),
	}
}

func DeserializeOutput(out *pb.Output) (*chain.Output) {
	return &chain.Output{
		NewOwner:     common.BytesToAddress(out.NewOwner),
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
