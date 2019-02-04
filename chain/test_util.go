package chain

import (
	"math/rand"
	"math/big"
	"github.com/ethereum/go-ethereum/common"
)

func RandomInput() *Input {
	return &Input{
		Output: Output{
			DepositNonce: big.NewInt(rand.Int63()),
			Owner: RandomAddress(),
			Denom: Zero(),
		},
		BlkNum: rand.Uint64(),
		TxIdx:  rand.Uint32(),
		OutIdx: 1,
	}
}

func RandomSig() []byte {
	size := 32
	result := make([]byte, size)
	rand.Read(result)
	return result
}

func RandomConfirmationSig() [65]byte {
	result := [65]byte{}
	rand.Read(result[:])
	return result
}

func RandomOutput() *Output {
	result := &Output{}
	result.Denom = big.NewInt(rand.Int63())
	result.DepositNonce = big.NewInt(0)
	buf := make([]byte, 20)
	rand.Read(buf)
	for i := range result.Owner {
		result.Owner[i] = buf[i]
	}
	return result
}

func RandomAddress() common.Address {
	buf := make([]byte, 20)
	rand.Read(buf)
	return common.BytesToAddress(buf)
}
