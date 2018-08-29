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

func extractAmount(tx *chain.Transaction, addr *common.Address) *big.Int {
    return tx.OutputFor(addr).Amount
}

func earnPrefixKey(addr *common.Address) []byte {
    return prefixKey(earnKeyPrefix, util.AddressToHex(addr))
}

func spendPrefixKey(addr *common.Address) []byte {
    return prefixKey(spendKeyPrefix, util.AddressToHex(addr))
}

func blkNumHashkey(blkNum uint64, hexHash string) []byte {
    return txPrefixKey("blkNum", strconv.FormatUint(blkNum, 10), "hash", hexHash)
}

func blkNumTxIdxKey(blkNum uint64, txIdx uint32) []byte {
    return txPrefixKey("blkNum", strconv.FormatUint(blkNum, 10), "txIdx", strconv.FormatUint(uint64(txIdx), 10))
}

func invalidPrefixKey(parts ...string) []byte {
    return prefixKey(invalidKeyPrefix, parts...)
}

func txPrefixKey(parts ...string) []byte {
    return prefixKey(txKeyPrefix, parts...)
}

func rlpMerkleTree(accepted []chain.Transaction) util.MerkleTree {
    hashables := make([]util.RLPHashable, len(accepted))

    for i := range accepted {
        txPtr := &accepted[i]
        hashables[i] = util.RLPHashable(txPtr)
    }

    merkle := util.TreeFromRLPItems(hashables)
    return merkle
}

func spend(addr *common.Address, input *chain.Input) []byte {
    blkNum := fmt.Sprintf("%d", input.BlkNum)
    txIdx  := fmt.Sprintf("%d", input.TxIdx)
    outIdx := fmt.Sprintf("%d", input.OutIdx)
    return prefixKey(spendKeyPrefix, util.AddressToHex(addr), blkNum, txIdx, outIdx)
}

func earn(addr *common.Address, tx chain.Transaction, outputIdx uint8) []byte {
    blkNum := fmt.Sprintf("%d", tx.BlkNum)
    txIdx  := fmt.Sprintf("%d", tx.TxIdx)
    outIdx := fmt.Sprintf("%d", outputIdx)
    return prefixKey(earnKeyPrefix, util.AddressToHex(addr), blkNum, txIdx, outIdx)
}

func parseSuffix(key []byte) (*common.Address, uint64, uint32, uint8, error) {
    parts := strings.Split(string(key), "::")
    if len(parts) != 4 {
        return nil, 0, 0, 0, errors.New("Invalid format key")
    }

    addr := common.BytesToAddress([]byte(parts[0]))

    blkNum, err := strconv.ParseUint(parts[1], 10, 64)
    if err != nil {
        return nil, 0, 0, 0, err
    }

    txIdx, err := strconv.ParseUint(parts[2], 10, 32)
    if err != nil {
        return nil, 0, 0, 0, err
    }

    outIdx, err := strconv.ParseUint(parts[3], 10, 32)
    if err != nil {
        return nil, 0, 0, 0, err
    }

    return &addr, blkNum, uint32(txIdx), uint8(outIdx), nil
}