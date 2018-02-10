package util

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/keybase/go-codec/codec"
	"log"
	"math/big"
	"reflect"
)

type BigIntExtension struct{}

func (BigIntExtension) ConvertExt(v interface{}) interface{} {
	switch v2 := v.(type) {
	case *big.Int:
		return "0x" + v2.Text(16)
	default:
		log.Panicf("unsupported format for BigInt conversion: expected *big.Int, got %T", v)
		return nil
	}
}

func (BigIntExtension) UpdateExt(dst interface{}, src interface{}) {
	b := dst.(*big.Int)

	switch v2 := src.(type) {
	case string:
		inst, ok := big.NewInt(0).SetString(v2, 0)

		if !ok {
			log.Panicf("failed to decode BigInt")
		}

		*b = *inst
	default:
		log.Panicf("unsupposed format for BigInt conversion: expected string, got %T", src)
	}
}

type HashExtension struct{}

func (HashExtension) ConvertExt(v interface{}) interface{} {
	switch v2 := v.(type) {
	case Hash:
		return common.ToHex(v2)
	default:
		log.Panicf("unsupported format for Hash conversion: expected Hash, got %T", v)
		return nil
	}
}

func (HashExtension) UpdateExt(dst interface{}, src interface{}) {
	b := dst.(*Hash)

	switch v2 := src.(type) {
	case string:
		data := common.FromHex(v2)
		*b = data
	default:
		log.Panicf("unsupposed format for Hash conversion: expected string, got %T", src)
	}
}

func PatchedCBORHandle() *codec.CborHandle {
	var bigExt BigIntExtension
	var hashExt HashExtension
	hdl := new(codec.CborHandle)
	hdl.SetInterfaceExt(reflect.TypeOf(big.Int{}), 0, bigExt)
	hdl.SetInterfaceExt(reflect.TypeOf(Hash{}), 1, hashExt)
	return hdl
}
