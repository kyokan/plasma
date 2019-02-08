package db

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/util"
	"github.com/pkg/errors"
)

const txKeyPrefix = "tx"
const earnKeyPrefix = "earn"
const spendKeyPrefix = "spend"
const spendExitKeyPrefix = "spend_exit"
const merkleKeyPrefix = "merkle"
const blockKeyPrefix = "blk"
const blockMetaKeyPrefix = "blkmeta"
const blockFees = "blk_fees"
const blockFeesExit = "blk_fees_exit"
const authSigPrefix = "auth_sig"
const depositPrefix = "deposit_nonce"
const latestKey = "LATEST_BLOCK"
const latestDepositIdxKey = "LATEST_DEPOSIT_IDX"
const latestTxExitIdxKey = "LATEST_TRANSACTION_EXIT_IDX"
const latestDepExitIdxKey = "LATEST_DEPOSIT_EXIT_IDX"
const invalidKeyPrefix = "invalid"
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

func blockFeesExitKey(number uint64) []byte {
	return prefixKey(blockFeesExit, strconv.FormatUint(number, 10))
}

func extractAmount(tx *chain.Transaction, addr *common.Address) *big.Int {
	return tx.OutputFor(addr).Denom
}

func earnPrefixKey(addr *common.Address) []byte {
	return prefixKey(earnKeyPrefix, util.AddressToHex(addr))
}

func spendPrefixKey(addr *common.Address) []byte {
	return prefixKey(spendKeyPrefix, util.AddressToHex(addr))
}

func blkNumHashkey(blkNum *big.Int, hexHash string) []byte {
	return txPrefixKey("blkNum", blkNum.String(), "hash", hexHash)
}

func blkNumTxIdxKey(blkNum uint64, txIdx uint32) []byte {
	return txPrefixKey("blkNum", strconv.FormatUint(uint64(blkNum), 10), "txIdx", strconv.FormatUint(uint64(txIdx), 10))
}

func blkNumTxIdxAuthSigKey(blkNum uint64, txIdx uint32) []byte {
	return prefixKey(authSigPrefix, string(blkNumTxIdxKey(blkNum, txIdx)))
}

// Used to lookup transaction associated with a deposit
// for handling deposit exit
func depositKey(confirmed *chain.ConfirmedTransaction) []byte {
	tx := &confirmed.Transaction
	return prefixKey(depositPrefix, tx.Output0.DepositNonce.String(), tx.BlkNum.String(), tx.TxIdx.String())
}

func depositPrefixKey(nonce *big.Int) []byte {
	return prefixKey(depositPrefix, nonce.String())
}

func invalidPrefixKey(parts ...string) []byte {
	return prefixKey(invalidKeyPrefix, parts...)
}

func txPrefixKey(parts ...string) []byte {
	return prefixKey(txKeyPrefix, parts...)
}

func spend(addr *common.Address, input *chain.Input) []byte {
	return prefixKey(
		spendKeyPrefix,
		util.AddressToHex(addr),
		strconv.FormatUint(input.BlkNum, 10),
		strconv.FormatUint(uint64(input.TxIdx), 10),
		strconv.FormatUint(uint64(input.OutIdx), 10),
		input.DepositNonce.Text(10),
	)
}

func spendExit(addr *common.Address, input *chain.Input) []byte {
	return prefixKey(
		spendExitKeyPrefix,
		util.AddressToHex(addr),
		strconv.FormatUint(input.BlkNum, 10),
		strconv.FormatUint(uint64(input.TxIdx), 10),
		strconv.FormatUint(uint64(input.OutIdx), 10),
		input.DepositNonce.Text(10),
	)
}

func earn(addr *common.Address, confirmed chain.ConfirmedTransaction, outputIdx uint8) []byte {
	tx := &confirmed.Transaction
	return prefixKey(earnKeyPrefix, util.AddressToHex(addr), tx.BlkNum.String(), tx.TxIdx.String(), strconv.FormatUint(uint64(outputIdx), 10))
}

func parseSuffix(key []byte) (*common.Address, *big.Int, *big.Int, *big.Int, error) {
	parts := strings.Split(string(key), "::")
	if len(parts) != 4 {
		return nil, nil, nil, nil, errors.New("Invalid format key")
	}

	addr := common.BytesToAddress([]byte(parts[0]))

	blkNum, success := new(big.Int).SetString(parts[1], 10)
	if success == false {
		return nil, nil, nil, nil, errors.New("Failed to parse block number")
	}

	txIdx, success := new(big.Int).SetString(parts[2], 10)
	if success == false {
		return nil, nil, nil, nil, errors.New("Failed to parse transaction index")
	}

	outIdx, success := new(big.Int).SetString(parts[3], 10)
	if success == false {
		return nil, nil, nil, nil, errors.New("Failed to parse output index")
	}

	return &addr, blkNum, txIdx, outIdx, nil
}
