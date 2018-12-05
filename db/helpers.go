package db

import (
    "fmt"
    "math/big"
    "strconv"
    "strings"

    "github.com/ethereum/go-ethereum/common"
    "github.com/kyokan/plasma/chain"
    "github.com/kyokan/plasma/util"
    "github.com/pkg/errors"
)

const txKeyPrefix         = "tx"
const earnKeyPrefix       = "earn"
const spendKeyPrefix      = "spend"
const merkleKeyPrefix     = "merkle"
const blockKeyPrefix      = "blk"
const blockMetaKeyPrefix  = "blkmeta"
const latestKey           = "LATEST_BLOCK"
const latestDepositIdxKey = "LATEST_DEPOSIT_IDX"
const latestExitIdxKey    = "LATEST_EXIT_IDX"
const invalidKeyPrefix    = "invalid"

// TODO: Read this from configuration
const blockSize           = uint32(100)

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

func blkNumTxIdxKey(blkNum, txIdx *big.Int) []byte {
    return txPrefixKey("blkNum", blkNum.String(), "txIdx", txIdx.String())
}

func invalidPrefixKey(parts ...string) []byte {
    return prefixKey(invalidKeyPrefix, parts...)
}

func txPrefixKey(parts ...string) []byte {
    return prefixKey(txKeyPrefix, parts...)
}

func spend(addr *common.Address, input *chain.Input) []byte {
    blkNum := fmt.Sprintf("%d", input.BlkNum)
    txIdx  := fmt.Sprintf("%d", input.TxIdx)
    outIdx := fmt.Sprintf("%d", input.OutIdx)
    return prefixKey(spendKeyPrefix, util.AddressToHex(addr), blkNum, txIdx, outIdx)
}

func earn(addr *common.Address, tx chain.Transaction, outputIdx *big.Int) []byte {
    return prefixKey(earnKeyPrefix, util.AddressToHex(addr), tx.BlkNum.String(), tx.TxIdx.String(), outputIdx.String())
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