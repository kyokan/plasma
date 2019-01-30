package chain

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/util"
	"math/big"
)

type Input struct {
	Output
	BlkNum *big.Int
	TxIdx  *big.Int
	OutIdx *big.Int
}

type rlpInput struct {
	BlkNum       *UInt256
	TxIdx        *UInt256
	OutIdx       *UInt256
	DepositNonce *UInt256
	Owner        common.Address
}

func NewInput(blkNum, txIdx, outIdx, depositNonce *big.Int, owner common.Address) *Input {
	return &Input{
		Output: Output{
			DepositNonce: depositNonce,
			Owner:        owner,
			Denom:        Zero(),
		},
		BlkNum: blkNum,
		TxIdx:  txIdx,
		OutIdx: outIdx,
	}
}

func ZeroInput() *Input {
	var address common.Address
	return NewInput(Zero(), Zero(), Zero(), Zero(), address)
}

func (in *Input) IsZeroInput() bool {
	return in == nil ||
			(in.BlkNum.Cmp(Zero()) == 0 &&
					in.TxIdx.Cmp(Zero()) == 0 &&
					in.OutIdx.Cmp(Zero()) == 0 &&
					in.DepositNonce.Cmp(Zero()) == 0)
}

func (in *Input) RLPHash(hasher util.Hasher) util.Hash {
	var itf rlpInput
	if in != nil {
		itf.BlkNum = NewUint256(in.BlkNum)
		itf.TxIdx = NewUint256(in.TxIdx)
		itf.OutIdx = NewUint256(in.OutIdx)
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
