package chain

import (
	"bytes"
	"encoding/binary"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/kyokan/plasma/util"
)

type Flow struct {
	BlkNum uint64
	TxIdx  uint32
	OutIdx uint8
	Hash   util.Hash
}

func NewFlow(blkNum uint64, txIdx uint32, outIdx uint8) *Flow {
	flow := &Flow{
		BlkNum: blkNum,
		TxIdx:  txIdx,
		OutIdx: outIdx,
	}
	hashFlow(flow)
	return flow
}

func hashFlow(flow *Flow) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, flow.BlkNum)
	binary.Write(buf, binary.BigEndian, flow.TxIdx)
	binary.Write(buf, binary.BigEndian, flow.OutIdx)
	digest := sha3.Sum256(buf.Bytes())
	flow.Hash = digest[:]
}
