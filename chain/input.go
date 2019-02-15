package chain

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/util"
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
	Owner        common.Address
}

type rlpInput struct {
	BlkNum       *UInt256
	TxIdx        *UInt256
	OutIdx       *UInt256
	DepositNonce *UInt256
	Owner        common.Address
}

func NewInput(blkNum uint64, txIdx uint32, outIdx uint8, depositNonce *big.Int, owner common.Address) *Input {
	return &Input{
		DepositNonce: depositNonce,
		Owner:        owner,
		BlockNum:     blkNum,
		TxIdx:        txIdx,
		OutIdx:       outIdx,
	}
}

func ZeroInput() *Input {
	var address common.Address
	return NewInput(0, 0, 0, Zero(), address)
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

func (in *Input) RLPHash(hasher util.Hasher) util.Hash {
	var itf rlpInput
	if in != nil {
		itf.BlkNum = NewUint256(new(big.Int).SetUint64(in.BlockNum))
		itf.TxIdx = NewUint256(big.NewInt(int64(in.TxIdx)))
		itf.OutIdx = NewUint256(big.NewInt(int64(in.OutIdx)))
		itf.DepositNonce = NewUint256(in.DepositNonce)
		itf.Owner = in.Owner
	} else {
		itf.BlkNum = NewUint256(nil)
		itf.TxIdx = NewUint256(nil)
		itf.OutIdx = NewUint256(nil)
		itf.DepositNonce = NewUint256(nil)
	}
	encoded, _ := rlp.EncodeToBytes(&itf)
	return hasher(encoded)
}

func (in *Input) SignatureHash() util.Hash {
	return in.RLPHash(util.Keccak256)
}

func (in *Input) Proto() (*pb.Input) {
	owner := make([]byte, len(in.Owner), len(in.Owner))
	copy(owner, in.Owner[:])

	return &pb.Input{
		BlockNum:     in.BlockNum,
		TxIdx:        in.TxIdx,
		OutIdx:       uint32(in.OutIdx),
		DepositNonce: rpc.SerializeBig(in.DepositNonce),
		Owner:        owner,
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
