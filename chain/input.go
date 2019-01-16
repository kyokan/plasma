package chain

import (
	"bytes"
	"encoding/binary"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/util"
	"math/big"
	"sync"
)

// JSON tags needed for test fixtures
type Input struct {
	BlkNum       *big.Int `json:"BlkNum"`
	TxIdx        *big.Int `json:"TxIdx"`
	OutIdx       *big.Int `json:"OutIdx"`
	DepositNonce *big.Int
	Owner        common.Address
}

func NewInput(blkNum, txIdx, outIdx, depositNonce *big.Int, owner common.Address) *Input {
	return &Input{
		BlkNum: blkNum,
		TxIdx:  txIdx,
		OutIdx: outIdx,
		DepositNonce: depositNonce,
		Owner:  owner,
	}
}

func NewInputFromTransaction(tx Transaction, outIdx int64) *Input {
	return &Input{
		BlkNum: tx.BlkNum,
		TxIdx:  tx.TxIdx,
		OutIdx: big.NewInt(outIdx),
		DepositNonce: Zero(),
	}
}

func ZeroInput() *Input {
	var address common.Address
	return NewInput(Zero(), Zero(), Zero(), Zero(), address)
}

func (in *Input) IsZeroInput() bool {
	return in.BlkNum.Cmp(Zero()) == 0 &&
			in.TxIdx.Cmp(Zero()) == 0 &&
			in.OutIdx.Cmp(Zero()) == 0 &&
			in.DepositNonce.Cmp(Zero()) == 0
}

func (in *Input) Hash() util.Hash {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, in.BlkNum)
	binary.Write(buf, binary.BigEndian, in.TxIdx)
	binary.Write(buf, binary.BigEndian, in.OutIdx)
	binary.Write(buf, binary.BigEndian, in.DepositNonce)
	binary.Write(buf, binary.BigEndian, in.Owner)
	digest := util.DoHash(buf.Bytes())
	return digest
}

var zero, one *big.Int
var once sync.Once

func initialize() {
	zero = big.NewInt(0)
	one  = big.NewInt(1)
}

func Zero() *big.Int {
	once.Do(initialize)
	return zero
}

func One() *big.Int {
	once.Do(initialize)
	return one
}

