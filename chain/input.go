package chain

import (
	"bytes"
	"encoding/binary"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/util"
	"math/big"
)

// JSON tags needed for test fixtures
type Input struct {
	Output
	BlkNum       *big.Int `json:"BlkNum"`
	TxIdx        *big.Int `json:"TxIdx"`
	OutIdx       *big.Int `json:"OutIdx"`
}

func NewInput(blkNum, txIdx, outIdx, depositNonce *big.Int, owner common.Address) *Input {
	return &Input{
		Output: Output{
			DepositNonce: depositNonce,
			Owner:  owner,
			Denom: Zero(),
		},
		BlkNum: blkNum,
		TxIdx:  txIdx,
		OutIdx: outIdx,
	}
}

func NewInputFromTransaction(tx Transaction, outIdx int64) *Input {
	return &Input{
		Output: Output{
			DepositNonce: Zero(),
		},
		BlkNum: tx.BlkNum,
		TxIdx:  tx.TxIdx,
		OutIdx: big.NewInt(outIdx),
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

func (in *Input) Hash() util.Hash {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, in.BlkNum)
	binary.Write(buf, binary.BigEndian, in.TxIdx)
	binary.Write(buf, binary.BigEndian, in.OutIdx)
	binary.Write(buf, binary.BigEndian, in.DepositNonce)
	binary.Write(buf, binary.BigEndian, in.Owner)
	digest := util.Keccak256(buf.Bytes())
	return digest
}

type rlpInputHelper struct {
	BlkNum       *UInt256
	TxIdx        *UInt256
	OutIdx       *UInt256
	DepositNonce *UInt256
	Owner        common.Address
}

func (in *Input) SignatureHash() util.Hash {
	var itf rlpInputHelper
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
	return util.Keccak256(encoded)
}

