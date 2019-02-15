package chain

import (
				"math/big"
	"github.com/kyokan/plasma/rpc/pb"
	"math"
	"github.com/pkg/errors"
	"github.com/kyokan/plasma/rpc"
)

type Input struct {
	BlockNum uint64
	TxIdx    uint32
	OutIdx   uint8
	DepositNonce *big.Int
}

func NewInput(blkNum uint64, txIdx uint32, outIdx uint8, depositNonce *big.Int) *Input {
	return &Input{
		DepositNonce: depositNonce,
		BlockNum:     blkNum,
		TxIdx:        txIdx,
		OutIdx:       outIdx,
	}
}

func ZeroInput() *Input {
	return NewInput(0, 0, 0, Zero())
}

func (in *Input) IsZero() bool {
	return in == nil ||
			(in.BlockNum == 0 &&
					in.TxIdx == 0 &&
					in.OutIdx == 0 &&
					in.DepositNonce.Cmp(Zero()) == 0)
}

func (in *Input) IsDeposit() bool {
	return in.BlockNum == 0 &&
			in.TxIdx == 0 &&
			in.OutIdx == 0 &&
			in.DepositNonce.Cmp(Zero()) != 0
}

func (in *Input) Proto() (*pb.Input) {
	return &pb.Input{
		BlockNum:     in.BlockNum,
		TxIdx:        in.TxIdx,
		OutIdx:       uint32(in.OutIdx),
		DepositNonce: rpc.SerializeBig(in.DepositNonce),
	}
}

func InputFromProto(protoIn *pb.Input) (*Input, error) {
	if protoIn == nil {
		return nil, errors.New("input cannot be nil")
	}

	if protoIn.OutIdx > math.MaxUint8 {
		return nil, errors.New("outIdx too large")
	}

	in := &Input{}
	in.BlockNum = protoIn.BlockNum
	in.TxIdx = protoIn.TxIdx
	in.OutIdx = uint8(protoIn.OutIdx)
	in.DepositNonce = rpc.DeserializeBig(protoIn.DepositNonce)
	return in, nil
}
