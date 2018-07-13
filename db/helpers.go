package db

import (
    "math/big"
    "strconv"

    "github.com/ethereum/go-ethereum/common"
    "github.com/kyokan/plasma/chain"
    "github.com/kyokan/plasma/util"
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

func earnKey(addr *common.Address) []byte {
    return prefixKey(earnKeyPrefix, util.AddressToHex(addr))
}

func spendKey(addr *common.Address) []byte {
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