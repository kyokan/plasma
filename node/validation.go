package node

import (
	"fmt"

	"github.com/kyokan/plasma/chain"
)

func EnsureNoDoubleSpend(txs []chain.Transaction) (okTxs []chain.Transaction, rejections []chain.Transaction) {
	if len(txs) == 1 {
		return txs, nil
	}

	used := make(map[string][]*chain.Transaction)
	var deposits []chain.Transaction

	for _, tx := range txs {
		if tx.IsDeposit() {
			deposits = append(deposits, tx)
		}

		txPtr := &tx
		keys := txToKeys(txPtr)

		for _, k := range keys {
			if _, exists := used[k]; exists {
				used[k] = append(used[k], &tx)
				continue
			}

			used[k] = []*chain.Transaction{&tx}
		}
	}

	output := make(map[*chain.Transaction]uint16)

	for _, txList := range used {
		for _, tx := range txList {
			if count, exists := output[tx]; exists {
				output[tx] = count + 1
			} else {
				output[tx] = 1
			}
		}
	}

	var ret []chain.Transaction
	var rej []chain.Transaction

	for tx, count := range output {
		if count > 1 {
			rej = append(rej, *tx)
		} else {
			ret = append(ret, *tx)
		}
	}

	return ret, rej
}

func FindMatchingInputs(tx *chain.Transaction, txs []chain.Transaction) (rejections []chain.Transaction) {
	usedKey0 := fmt.Sprintf("%d::%d::%d", tx.BlkNum, tx.TxIdx, 0)
	usedKey1 := fmt.Sprintf("%d::%d::%d", tx.BlkNum, tx.TxIdx, 1)

	var used []chain.Transaction

	for _, currTx := range txs {
		if currTx.IsDeposit() {
			continue
		}

		keys := txToKeys(&currTx)

		for _, k := range keys {
			if k == usedKey0 || k == usedKey1 {
				used = append(used, currTx)
			}
		}
	}

	return used
}

func txToKeys(tx *chain.Transaction) []string {
	if tx.IsDeposit() {
		return nil
	}

	keys := make([]string, 2)
	keys = append(keys, fmt.Sprintf("%d::%d::%d", tx.Input0.BlkNum, tx.Input0.TxIdx, tx.Input0.OutIdx))

	if !tx.Input1.IsZeroInput() {
		keys = append(keys, fmt.Sprintf("%d::%d::%d", tx.Input1.BlkNum, tx.Input1.TxIdx, tx.Input1.OutIdx))
	}

	return keys
}
