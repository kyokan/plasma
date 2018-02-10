package chain

import (
	"bytes"
	"encoding/binary"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/kyokan/plasma/util"
)

type Input struct {
	BlkNum uint64
	TxIdx  uint32
	OutIdx uint8
}

func ZeroInput() *Input {
	return &Input{
		BlkNum: 0,
		TxIdx:  0,
		OutIdx: 0,
	}
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
	digest := sha3.Sum256(buf.Bytes())
	return digest[:]
}
