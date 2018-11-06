package chain

import (
	"bytes"
	"encoding/binary"
	"github.com/kyokan/plasma/util"
)

// JSON tags needed for test fixtures
type Input struct {
	BlkNum uint64 `json:"BlkNum"`
	TxIdx  uint32 `json:"TxIdx"`
	OutIdx uint8  `json:"OutIdx"`
}

func NewInput(blkNum uint64, txIdx uint32, outIdx uint8) *Input {
	return &Input{
		BlkNum: blkNum,
		TxIdx:  txIdx,
		OutIdx: outIdx,
	}
}

func ZeroInput() *Input {
	return NewInput(0, 0, 0)
}

func (in *Input) IsZeroInput() bool {
	return in.BlkNum == 0 &&
		in.TxIdx == 0 &&
		in.OutIdx == 0
}

func (in *Input) Hash() util.Hash {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, in.BlkNum)
	binary.Write(buf, binary.BigEndian, in.TxIdx)
	binary.Write(buf, binary.BigEndian, in.OutIdx)
	digest := util.DoHash(buf.Bytes())
	return digest
}
