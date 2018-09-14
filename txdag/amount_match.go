package txdag

import (
    "math/big"
    "sort"

    "github.com/ethereum/go-ethereum/common"
    "github.com/pkg/errors"
        "github.com/kyokan/plasma/chain"
)

type OutputSortHelper struct {
    Position  int
    Amount    *big.Int
}

// FindBestUTXOs Finds (at most two) UXTOs to match an amount.
func FindBestUTXOs(from, to common.Address, amount *big.Int, txs []chain.Transaction) (*chain.Transaction, error) {
    if len(txs) == 0 {
        return nil, errors.New("no suitable UTXOs found")
    }
    outputs := make([]OutputSortHelper, 0, len(txs))
    for pos, tx := range txs {
        output := tx.OutputFor(&from) // this call may panic
        if amount.Cmp(output.Amount) == 0 {
            // Found exact match
            return PrepareSendTransaction(from, to, amount, []chain.Transaction{txs[pos]})
        }
        outputs = append(outputs, OutputSortHelper{Position: pos, Amount: output.Amount})
    }
    less := func(i, j int) bool { // return outputs[i] < outputs[j]
        lhs := outputs[i].Amount
        rhs := outputs[j].Amount
        return lhs.Cmp(rhs) == -1
    }
    sort.Slice(outputs, less)
    // Amount is less the minimum element, no need to do anything else
    min := outputs[0]
    if min.Amount.Cmp(amount) == 1 { // min > amount
        return PrepareSendTransaction(from, to, amount, []chain.Transaction{txs[min.Position]})
    }
    leftBound := int(0)
    rightBound := len(outputs) - 1
    lhs := -1
    rhs := -1
    for ; leftBound < rightBound;  {
        sum := big.NewInt(0)
        sum.Add(outputs[leftBound].Amount, outputs[rightBound].Amount)
        cmp := sum.Cmp(amount)
        if cmp == 0 { // sum == amount
            break
        }
        if cmp == -1 { // sum < amount
            leftBound++
            continue
        }
        // keep track of last sum greater than amount
        lhs = leftBound
        rhs = rightBound
        rightBound-- // sum > amount
    }
    if leftBound < rightBound { // Found two outputs that sum up to amount
        first := outputs[leftBound].Position
        second := outputs[rightBound].Position
        return PrepareSendTransaction(from, to, amount, []chain.Transaction{txs[first], txs[second]})
    }
    if lhs >= 0 && rhs >= 0 { // smallest sum that's greater than amount
        first := outputs[lhs].Position
        second := outputs[rhs].Position
        return PrepareSendTransaction(from, to, amount, []chain.Transaction{txs[first], txs[second]})
    }
    return nil, errors.New("no suitable UTXOs found")
}

func PrepareSendTransaction(from, to common.Address, amount *big.Int, utxoTxs []chain.Transaction) (*chain.Transaction, error) {
    var input1 *chain.Input
    var output1 *chain.Output
    totalAmount := big.NewInt(0)

    if len(utxoTxs) == 1 {
        input1 = chain.ZeroInput()

        utxo := utxoTxs[0].OutputFor(&from)

        if utxo == nil {
            return nil, errors.New("expected a UTXO")
        }

        totalAmount.Set(utxo.Amount)
    } else {
        input1 = &chain.Input{
            BlkNum: utxoTxs[1].BlkNum,
            TxIdx:  utxoTxs[1].TxIdx,
            OutIdx: utxoTxs[1].OutputIndexFor(&from),
        }

        totalAmount = totalAmount.Add(utxoTxs[0].OutputFor(&from).Amount, utxoTxs[1].OutputFor(&from).Amount)
    }
    if totalAmount.Cmp(amount) == 1 { // totalAmount > amount
        output1 = &chain.Output{
            NewOwner: from,
            Amount:   big.NewInt(0).Sub(totalAmount, amount),
        }
    } else {
        output1 = chain.ZeroOutput()
    }

    tx := chain.Transaction{
        Input0: &chain.Input{
            BlkNum: utxoTxs[0].BlkNum,
            TxIdx:  utxoTxs[0].TxIdx,
            OutIdx: utxoTxs[0].OutputIndexFor(&from),
        },
        Input1: input1,
        Output0: &chain.Output{
            NewOwner: to,
            Amount:   amount,
        },
        Output1: output1,
        Fee:     big.NewInt(0),
    }
    if tx.Input1.IsZeroInput() == false {
        //Input1 is valid, set the signature (note that signature is the same)
        tx.Sig1 = tx.Sig0
    }
    return &tx, nil
}
