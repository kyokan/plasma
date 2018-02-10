package chain

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/keybase/go-codec/codec"
	"github.com/kyokan/plasma/util"
)

type Flow struct {
	BlkNum uint64
	TxIdx  uint32
	OutIdx uint8
	Hash   util.Hash
}

func (f *Flow) ToCbor() ([]byte, error) {
	buf := new(bytes.Buffer)
	bw := bufio.NewWriter(buf)
	hdl := util.PatchedCBORHandle()
	enc := codec.NewEncoder(bw, hdl)
	err := enc.Encode(f)

	if err != nil {
		return nil, err
	}

	bw.Flush()

	return buf.Bytes(), nil
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

func FlowFromCbor(data []byte) (*Flow, error) {
	hdl := util.PatchedCBORHandle()
	dec := codec.NewDecoderBytes(data, hdl)
	ptr := &Flow{}
	err := dec.Decode(ptr)

	if err != nil {
		return nil, err
	}

	return ptr, nil
}

func hashFlow(flow *Flow) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, flow.BlkNum)
	binary.Write(buf, binary.BigEndian, flow.TxIdx)
	binary.Write(buf, binary.BigEndian, flow.OutIdx)
	digest := sha3.Sum256(buf.Bytes())
	flow.Hash = digest[:]
}
