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
	buf := new(bytes.Buffer)
	buf.Write([]byte(prefix))

	for _, part := range parts {
		buf.Write([]byte(keyPartsSeparator))
		buf.Write([]byte(part))
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
	txLess := func (lhs, rhs int) bool {
		if txs[lhs].Transaction.BlkNum.Cmp(txs[rhs].Transaction.BlkNum) == 0 {
			return txs[lhs].Transaction.TxIdx.Cmp(txs[rhs].Transaction.TxIdx) == -1
		}
		return txs[lhs].Transaction.BlkNum.Cmp(txs[rhs].Transaction.BlkNum) == -1
	}
	sort.Slice(txs, txLess)
}
