package db

import (
	"bytes"
	"fmt"
	"github.com/kyokan/plasma/chain"
	"log"
	"sort"
	"strconv"
)

const keyPartsSeparator = "::"

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
