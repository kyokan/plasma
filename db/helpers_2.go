package db

import (
	"github.com/kyokan/plasma/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"strings"
	"math/big"
)

const (
	txPrefix      = "tx"
	spendPrefix   = "spend"
	utxoPrefix    = "utxo"
	depositPrefix = "deposit"
)

func txByHashKey(hash string) []byte {
	return joinKey(txPrefix, "hash", hash)
}

func txByBlockNumTxIdxKey(blockNum uint64, txIdx uint32) []byte {
	return joinKey(txPrefix, "blockNumTxIdx", util.Uint642Str(blockNum), util.Uint322Str(txIdx))
}

func txByBlockNumIterKey(blockNum uint64) []byte {
	return joinKey(txPrefix, "blockNumTxIdx", util.Uint642Str(blockNum))
}

func utxoKey(addr common.Address, hash util.Hash, outIdx uint8) []byte {
	return joinKey(utxoPrefix, hexutil.Encode(addr[:]), hexutil.Encode(hash), util.Uint82Str(outIdx))
}

func utxoAddrIterKey(addr common.Address) []byte {
	return joinKey(utxoPrefix, hexutil.Encode(addr[:]))
}

func depositKey(nonce *big.Int) []byte {
	return joinKey(depositPrefix, util.Big2Str(nonce))
}

func utxoKeyParts(key string) (util.Hash, uint8) {
	parts := strings.Split(key, keyPartsSeparator)
	if len(parts) != 4 {
		panic("invalid UTXO key")
	}
	hash, err := hexutil.Decode(parts[2])
	if err != nil {
		panic("invalid tx hash")
	}
	outIdx, ok := util.Str2Uint8(parts[3])
	if !ok {
		panic("invalid output index")
	}
	return hash, outIdx
}

func spendByTxIdxKey(blockNum uint64, txIdx uint32, outIdx uint8) []byte {
	return joinKey(
		spendPrefix,
		"blockNumTxIdx",
		util.Uint642Str(blockNum),
		util.Uint322Str(txIdx),
		util.Uint82Str(outIdx),
	)
}
