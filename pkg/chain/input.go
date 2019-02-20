package chain

import (
	"math/big"
	"github.com/kyokan/plasma/pkg/rpc/pb"
		"github.com/pkg/errors"
	"github.com/kyokan/plasma/pkg/rpc"
	"github.com/kyokan/plasma/util"
	"encoding/json"
)

type Input struct {
	BlockNumber      uint64
	TransactionIndex uint32
	OutputIndex      uint8
	DepositNonce     *big.Int
}

type inputJSON struct {
	BlockNumber      uint64 `json:"blockNumber"`
	TransactionIndex uint32 `json:"transactionIndex"`
	OutputIndex      uint8  `json:"outputIndex"`
	DepositNonce     string `json:"depositNonce"`
}

func NewInput(blkNum uint64, txIdx uint32, outIdx uint8, depositNonce *big.Int) *Input {
	return &Input{
		DepositNonce:     depositNonce,
		BlockNumber:      blkNum,
		TransactionIndex: txIdx,
		OutputIndex:      outIdx,
	}
}

func ZeroInput() *Input {
	return NewInput(0, 0, 0, Zero())
}

func (in *Input) MarshalJSON() ([]byte, error) {
	jsonRep := &inputJSON{
		in.BlockNumber,
		in.TransactionIndex,
		in.OutputIndex,
		util.Big2Str(in.DepositNonce),
	}

	return json.Marshal(jsonRep)
}

func (in *Input) UnmarshalJSON(raw []byte) error {
	jsonRep := &inputJSON{}
	err := json.Unmarshal(raw, jsonRep)
	if err != nil {
		return err
	}
	nonce, err := util.Str2Big(jsonRep.DepositNonce)
	if err != nil {
		return err
	}
	in.BlockNumber = jsonRep.BlockNumber
	in.TransactionIndex = jsonRep.TransactionIndex
	in.OutputIndex = jsonRep.OutputIndex
	in.DepositNonce = nonce
	return nil
}

func (in *Input) IsZero() bool {
	return in == nil ||
			(in.BlockNumber == 0 &&
					in.TransactionIndex == 0 &&
					in.OutputIndex == 0 &&
					in.DepositNonce.Cmp(Zero()) == 0)
}

func (in *Input) IsDeposit() bool {
	return in.BlockNumber == 0 &&
			in.TransactionIndex == 0 &&
			in.OutputIndex == 0 &&
			in.DepositNonce.Cmp(Zero()) != 0
}

func (in *Input) Proto() (*pb.Input) {
	return &pb.Input{
		BlockNum:     in.BlockNumber,
		TxIdx:        in.TransactionIndex,
		OutIdx:       uint32(in.OutputIndex),
		DepositNonce: rpc.SerializeBig(in.DepositNonce),
	}
}

func InputFromProto(protoIn *pb.Input) (*Input, error) {
	if protoIn == nil {
		return nil, errors.New("input cannot be nil")
	}

	if protoIn.OutIdx > 1 {
		return nil, errors.New("outIdx too large")
	}

	in := &Input{}
	in.BlockNumber = protoIn.BlockNum
	in.TransactionIndex = protoIn.TxIdx
	in.OutputIndex = uint8(protoIn.OutIdx)
	in.DepositNonce = rpc.DeserializeBig(protoIn.DepositNonce)
	return in, nil
}
