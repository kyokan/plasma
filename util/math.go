package util

import "math/big"

func Add(a *big.Int, b int64) *big.Int {
	res := new(big.Int)
	res.Add(a, new(big.Int).SetInt64(b))
	return res
}

func Sub(a *big.Int, b int64) *big.Int {
	res := new(big.Int)
	res.Sub(a, new(big.Int).SetInt64(b))
	return res
}

func NewInt64(a int64) *big.Int {
	return new(big.Int).SetInt64(a)
}

func NewUint64(a uint64) *big.Int {
	return new(big.Int).SetUint64(a)
}

func NewUint32(a uint32) *big.Int {
	return new(big.Int).SetUint64(uint64(a))
}

func NewInt(a int) *big.Int {
	return new(big.Int).SetInt64(int64(a))
}
