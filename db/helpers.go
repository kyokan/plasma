package db

import (
	"math/big"
	"strconv"
		"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
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