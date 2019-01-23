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
func FindBestUTXOs(from, to common.Address, amount *big.Int, txs []chain.ConfirmedTransaction) (*chain.ConfirmedTransaction, error) {
    if len(txs) == 0 {
        return nil, errors.New("no suitable UTXOs found")
    }
    result := &chain.ConfirmedTransaction{}
    outputs := make([]OutputSortHelper, 0, len(txs))
    for pos, tx := range txs {
        output := tx.Transaction.OutputFor(&from) // this call may panic
        if amount.Cmp(output.Denom) == 0 {
            // Found exact match
            transaction, err := PrepareSendTransaction(from, to, amount, []chain.ConfirmedTransaction{txs[pos]})
            result.Transaction = *transaction
            return result, err
        }
        outputs = append(outputs, OutputSortHelper{Position: pos, Amount: output.Denom})
    }
    less := func(i, j int) bool { // return outputs[i] < outputs[j]
        lhs := outputs[i].Amount
        rhs := outputs[j].Amount
        return lhs.Cmp(rhs) == -1
    }
    sort.Slice(outputs, less)
    // Denom is less the minimum element, no need to do anything else
    min := outputs[0]
    if min.Amount.Cmp(amount) == 1 { // min > amount
        transaction, err := PrepareSendTransaction(from, to, amount, []chain.ConfirmedTransaction{txs[min.Position]})
        result.Transaction = *transaction
        return result, err
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
        transaction, err := PrepareSendTransaction(from, to, amount, []chain.ConfirmedTransaction{txs[first], txs[second]})
        result.Transaction = *transaction
        return result, err
    }
    if lhs >= 0 && rhs >= 0 { // smallest sum that's greater than amount
        first := outputs[lhs].Position
        second := outputs[rhs].Position
        transaction, err := PrepareSendTransaction(from, to, amount, []chain.ConfirmedTransaction{txs[first], txs[second]})
        result.Transaction = *transaction
        return result, err
    }
    return nil, errors.New("no suitable UTXOs found")
}

func PrepareSendTransaction(from, to common.Address, amount *big.Int, utxoTxs []chain.ConfirmedTransaction) (*chain.Transaction, error) {
    var input1 *chain.Input
    var output1 *chain.Output
    totalAmount := big.NewInt(0)

    if len(utxoTxs) == 1 {
        input1 = chain.ZeroInput()

        utxo := utxoTxs[0].Transaction.OutputFor(&from)

        if utxo == nil {
            return nil, errors.New("expected a UTXO")
        }

        totalAmount.Set(utxo.Denom)
    } else {

        input1 = &chain.Input{
            Output: chain.Output{},
            BlkNum: utxoTxs[1].Transaction.BlkNum,
            TxIdx:  utxoTxs[1].Transaction.TxIdx,
            OutIdx: utxoTxs[1].Transaction.OutputIndexFor(&from),
        }

        totalAmount = totalAmount.Add(utxoTxs[0].Transaction.OutputFor(&from).Denom, utxoTxs[1].Transaction.OutputFor(&from).Denom)
    }
    if totalAmount.Cmp(amount) == 1 { // totalAmount > amount
        output1 = &chain.Output{
            Owner: from,
            Denom: big.NewInt(0).Sub(totalAmount, amount),
        }
    } else {
        output1 = chain.ZeroOutput()
    }

    tx := chain.Transaction{
        Input0: &chain.Input{
            BlkNum: utxoTxs[0].Transaction.BlkNum,
            TxIdx:  utxoTxs[0].Transaction.TxIdx,
            OutIdx: utxoTxs[0].Transaction.OutputIndexFor(&from),
        },
        Input1: input1,
        Output0: &chain.Output{
            Owner: to,
            Denom: amount,
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
