package db

import (
	"math/big"
	"strconv"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/pkg/chain"
	"github.com/kyokan/plasma/util"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"strings"
	"bytes"
	"fmt"
	"log"
	"sort"
	"github.com/syndtr/goleveldb/leveldb"
	"path"
	"os/user"
)

const keyPartsSeparator = "::"
const merkleKeyPrefix = "merkle"
const blockKeyPrefix = "blk"
const blockMetaKeyPrefix = "blkmeta"
const exitPrefix = "exit"
const latestKey = "LATEST_BLOCK"
const latestDepositIdxKey = "LATEST_DEPOSIT_IDX"
const lastTxExitPollKey = "LATEST_TRANSACTION_EXIT_IDX"
const latestDepExitIdxKey = "LATEST_DEPOSIT_EXIT_IDX"
const lastSubmittedBlockKey = "LAST_SUBMITTED_BLOCK"

func CreateLevelStorage(location string) (*leveldb.DB, Storage, error) {
	loc := path.Join(location, "db")
	level, err := leveldb.OpenFile(loc, nil)
	if err != nil {
		return nil, nil, err
	}
	return level, NewLevelStorage(level), nil
}

func DefaultLevelLocation() string {
	usr, err := user.Current()

	if err != nil {
		log.Fatal(err)
	}

	return path.Join(usr.HomeDir, ".plasma")
}

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

func exitKey(blockNum uint64, txIdx uint32, outIdx uint8) []byte {
	return joinKey(
		exitPrefix,
		util.Uint642Str(blockNum),
		util.Uint322Str(txIdx),
		util.Uint82Str(outIdx),
	)
}

func prefixKey(prefix string, parts ...string) []byte {
	var args []string
	args = append(args, prefix)
	args = append(args, parts...)
	return joinKey(args...)
}

func joinKey(parts ...string) []byte {
	buf := new(bytes.Buffer)
	for i, part := range parts {
		buf.Write([]byte(part))

		if i != len(parts)-1 {
			buf.Write([]byte(keyPartsSeparator))
		}
	}
	return buf.Bytes()
}

func uint64ToBytes(i uint64) []byte {
	return []byte(fmt.Sprintf("%X", i))
}

func bytesToUint64(b []byte) uint64 {
	s := string(b)

	n, err := strconv.ParseUint(s, 16, 32)

	if err != nil {
		log.Fatalf("Failed to parse string as hex: %v", err)
	}

	return uint64(n)
}

func sortTransactions(txs []chain.ConfirmedTransaction) {
	txLess := func(lhs, rhs int) bool {
		if txs[lhs].Transaction.Body.BlockNumber == txs[rhs].Transaction.Body.BlockNumber {
			return txs[lhs].Transaction.Body.TransactionIndex < txs[rhs].Transaction.Body.TransactionIndex
		}
		return txs[lhs].Transaction.Body.BlockNumber < txs[rhs].Transaction.Body.BlockNumber
	}
	sort.Slice(txs, txLess)
}
