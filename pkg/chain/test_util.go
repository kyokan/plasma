package chain

import (
	"math/rand"
	"math/big"
	"github.com/ethereum/go-ethereum/common"
)

func RandomInput() *Input {
	return &Input{
		DepositNonce:     big.NewInt(rand.Int63()),
		BlockNumber:      rand.Uint64(),
		TransactionIndex: rand.Uint32(),
		OutputIndex:      1,
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
	result.Amount = big.NewInt(rand.Int63())
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
