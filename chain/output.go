package chain

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/kyokan/plasma/util"
	"math/big"
)

// JSON tags needed for test fixtures
type Output struct {
	NewOwner common.Address `json:"NewOwner"`
	Amount   *big.Int       `json:"NewOwner"`
}

func ZeroOutput() *Output {
	return &Output{
		NewOwner: common.BytesToAddress(make([]byte, 20, 20)),
		Amount:   big.NewInt(0),
	}
}

func (out *Output) IsZeroOutput() bool {
	addrBytes := out.NewOwner.Bytes()

	for _, v := range addrBytes {
		if v != 0 {
			return false
		}
	}

	return out.Amount.Cmp(big.NewInt(0)) == 0
}

func (out *Output) Hash() util.Hash {
	buf := new(bytes.Buffer)
	buf.Write(out.NewOwner.Bytes())
	buf.Write(out.Amount.Bytes())
	digest := sha3.Sum256(buf.Bytes())
	return digest[:]
}
