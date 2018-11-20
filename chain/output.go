package chain

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/util"
	"math/big"
)

// JSON tags needed for test fixtures
type Output struct {
	NewOwner     common.Address `json:"NewOwner"`
	Denom        *big.Int       `json:"Denom"`
	DepositNonce *big.Int		`json:"Nonce"`
}

func NewOutput(newOwner common.Address, amount *big.Int) *Output {
	return &Output{
		NewOwner: common.BytesToAddress(newOwner.Bytes()),
		Denom:    big.NewInt(amount.Int64()),
	}
}

func ZeroOutput() *Output {
	return &Output{
		NewOwner:     common.BytesToAddress(make([]byte, 20, 20)),
		Denom:        big.NewInt(0),
		DepositNonce: big.NewInt(0),
	}
}

func (out *Output) IsZeroOutput() bool {
	addrBytes := out.NewOwner.Bytes()

	for _, v := range addrBytes {
		if v != 0 {
			return false
		}
	}

	return out.Denom.Cmp(big.NewInt(0)) == 0 &&
		   out.DepositNonce.Cmp(big.NewInt(0)) == 0
}

func (out *Output) Hash() util.Hash {
	buf := new(bytes.Buffer)
	buf.Write(out.NewOwner.Bytes())
	buf.Write(out.Denom.Bytes())
	digest := util.DoHash(buf.Bytes())
	return digest
}
