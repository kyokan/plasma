package chain

import (
	"math/rand"
	"math/big"
	"github.com/ethereum/go-ethereum/common"
)

func RandomInput() *Input {
	return &Input{
		BlkNum: rand.Uint64(),
		TxIdx:  rand.Uint32(),
		OutIdx: uint8(rand.Uint32()),
	}
}

func RandomSig() []byte {
	size := 32
	result := make([]byte, size)
	rand.Read(result)
	return result
}

func RandomOutput() *Output {
	result := &Output{}
	result.Amount = big.NewInt(rand.Int63())
	buf := make([]byte, 20)
	rand.Read(buf)
	for i := range result.NewOwner {
		result.NewOwner[i] = buf[i]
	}
	return result
}

func RandomAddress() common.Address {
	buf := make([]byte, 20)
	rand.Read(buf)
	return common.BytesToAddress(buf)
}
