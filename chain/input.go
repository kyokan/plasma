package chain

import (
	"bytes"
	"encoding/binary"
	"github.com/kyokan/plasma/util"
	"math/big"
	"sync"
)

// JSON tags needed for test fixtures
type Input struct {
	BlkNum       *big.Int `json:"BlkNum"`
	TxIdx        *big.Int `json:"TxIdx"`
	OutIdx       *big.Int `json:"OutIdx"`
}

func NewInput(blkNum, txIdx, outIdx *big.Int) *Input {
	return &Input{
		BlkNum: blkNum,
		TxIdx:  txIdx,
		OutIdx: outIdx,
	}
}

func ZeroInput() *Input {
	zero := big.NewInt(0)
	return NewInput(zero, zero, zero)
}

func (in *Input) IsZeroInput() bool {
	return in.BlkNum == nil &&
			in.TxIdx == nil &&
			in.OutIdx == nil
}

func (in *Input) Hash() util.Hash {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, in.BlkNum)
	binary.Write(buf, binary.BigEndian, in.TxIdx)
	binary.Write(buf, binary.BigEndian, in.OutIdx)
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

