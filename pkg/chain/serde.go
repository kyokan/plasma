package chain

import (
	"math/big"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"encoding/json"
	)

type Signature [65]byte

func (s Signature) MarshalJSON() ([]byte, error) {
	str := hexutil.Encode(s[:])
	return json.Marshal(str)
}

func (s *Signature) UnmarshalJSON(in []byte) error {
	var sigStr string
	err := json.Unmarshal(in, &sigStr)
	if err != nil {
		return err
	}
	res, err := hexutil.Decode(sigStr)
	var sig Signature
	if err != nil {
		return err
	}
	copy(sig[:], res)
	*s = sig
	return nil
}

type UInt256 [32]byte

func NewUint256(i *big.Int) *UInt256 {
	var result UInt256
	if i != nil {
		bytes := i.Bytes()
		diff := len(result) - len(bytes)
		for i := 0; i != len(bytes); i++ {
			result[diff+i] = bytes[i]
		}
	}
	return &result
}

func (uint *UInt256) ToBig() *big.Int {
	result := big.NewInt(0)
	if uint != nil {
		firstNonZero := 0
		for ; firstNonZero != len(*uint); firstNonZero++ {
			if uint[firstNonZero] != 0 {
				break
			}
		}
		result.SetBytes((*uint)[firstNonZero:])
	}
	return result
}
