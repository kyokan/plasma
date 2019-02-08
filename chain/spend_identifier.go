package chain

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type SpendIdentifier struct {
	BlockNumber      uint64
	TransactionIndex uint32
	InputIndex       uint8
}

func (s *SpendIdentifier) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	blkNumBuf := make([]byte, 8, 8)
	binary.BigEndian.PutUint64(blkNumBuf, s.BlockNumber)
	buf.Write(blkNumBuf)
	txIdxBuf := make([]byte, 4, 4)
	binary.BigEndian.PutUint32(txIdxBuf, s.TransactionIndex)
	buf.Write(txIdxBuf)
	buf.WriteByte(byte(s.InputIndex))
	return buf.Bytes(), nil
}

func (s *SpendIdentifier) UnmarshalBinary(data []byte) (error) {
	if len(data) != 13 {
		return errors.New("invalid length")
	}

	s.BlockNumber = binary.BigEndian.Uint64(data[0:8])
	s.TransactionIndex = binary.BigEndian.Uint32(data[8:12])
	s.InputIndex = uint8(data[12])
	return nil
}