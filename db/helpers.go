package db

import (
	"math/big"
	"strconv"
		"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/util"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"strings"
)

const merkleKeyPrefix = "merkle"
const blockKeyPrefix = "blk"
const blockMetaKeyPrefix = "blkmeta"
const blockFees = "blk_fees"
const latestKey = "LATEST_BLOCK"
const latestDepositIdxKey = "LATEST_DEPOSIT_IDX"
const lastTxExitPollKey = "LATEST_TRANSACTION_EXIT_IDX"
const latestDepExitIdxKey = "LATEST_DEPOSIT_EXIT_IDX"
const lastSubmittedBlockKey = "LAST_SUBMITTED_BLOCK"

func merklePrefixKey(parts ...string) []byte {
	return prefixKey(merkleKeyPrefix, parts...)
}

func blockNumKey(num uint64) []byte {
	return blockPrefixKey(strconv.FormatUint(num, 10))
}

func blockPrefixKey(parts ...string) []byte {
	return prefixKey(blockKeyPrefix, parts...)
}

func blockMetaPrefixKey(number uint64) []byte {
	return prefixKey(blockMetaKeyPrefix, strconv.FormatUint(number, 10))
}

func blockFeesKey(number uint64) []byte {
	return prefixKey(blockFees, strconv.FormatUint(number, 10))
}

func extractAmount(tx *chain.ConfirmedTransaction, addr common.Address) *big.Int {
	return tx.Transaction.Body.OutputFor(&addr).Amount
}

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