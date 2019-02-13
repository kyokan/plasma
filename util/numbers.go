package util

import (
	"math/big"
	"math"
	"strconv"
)

var MaxUint64 = Uint642Big(math.MaxUint64)
var MaxUint32 = Uint322Big(math.MaxUint32)
var MaxUint8 = Uint82Big(math.MaxUint8)

func Uint642Big(n uint64) *big.Int {
	return new(big.Int).SetUint64(n)
}

func Uint322Big(n uint32) *big.Int {
	return big.NewInt(int64(n))
}

func Uint82Big(n uint8) *big.Int {
	return big.NewInt(int64(n))
}

func Big2Uint64(n *big.Int) uint64 {
	return n.Uint64()
}

func Big2Uint32(n *big.Int) uint32 {
	if n.Cmp(MaxUint32) == 1 {
		panic("overflow")
	}

	return uint32(n.Uint64())
}

func Big2Uint8(n *big.Int) uint8 {
	if n.Cmp(MaxUint8) == 1 {
		panic("overflow")
	}

	return uint8(n.Uint64())
}

func Uint642Str(n uint64) string {
	return strconv.FormatUint(n, 10)
}

func Uint322Str(n uint32) string {
	return Uint642Str(uint64(n))
}

func Str2Uint64(s string) (uint64, bool) {
	res, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return uint64(res), true
}

func Str2Uint32(s string) (uint32, bool) {
	res, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, false
	}
	return uint32(res), true
}
