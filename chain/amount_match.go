package chain

import (
    "math/big"
    "math/rand"

    "github.com/ethereum/go-ethereum/common"
    plasma_common "github.com/kyokan/plasma/common"
    "github.com/pkg/errors"
)

func FindBestUTXOs(from, to common.Address, amount *big.Int, txs []Transaction, client plasma_common.Client) (*Transaction, error) {

    // similar algo to the one Bitcoin uses: https://bitcoin.stackexchange.com/questions/1077/what-is-the-coin-selection-algorithm

    candidates := make([]Transaction, 0, 2)
    // pass 1: find UTXOs that exactly match amount
    for _, tx := range txs {
        utxo := tx.OutputFor(&from)

        if utxo.Amount.Cmp(amount) == 0 {
            candidates = append(candidates, tx)
            break
        }
    }
    if len(candidates) > 0 {
        return PrepareSendTransaction(from, to, amount, candidates, client)
    }
    // pass 2: sum of any two UTXOs exactly matches amount
    for i, ltx := range txs {
        for _, rtx := range txs[i:] {
            total := big.NewInt(0)
            total = total.Add(ltx.OutputFor(&from).Amount, rtx.OutputFor(&from).Amount)

            if total.Cmp(amount) == 0 {
                candidates = append(candidates, ltx)
                candidates = append(candidates, rtx)
                break
            }
        }
        if len(candidates) > 0 {
            break
        }
    }
    if len(candidates) > 0 {
        return PrepareSendTransaction(from, to, amount, candidates, client)
    }

    // pass 3: find smallest UTXO larger than amount
    var ret *Transaction
    var closestAmount *big.Int

    for _, tx := range txs {
        utxo := tx.OutputFor(&from)

        if utxo.Amount.Cmp(amount) == -1 {
            continue
        }

        if closestAmount == nil {
            closestAmount = utxo.Amount
            ret = &tx
        }

        if closestAmount.Cmp(utxo.Amount) == -1 {
            continue
        }

        closestAmount = utxo.Amount
        ret = &tx
    }

    if ret != nil {
        candidates = append(candidates, *ret)
        return PrepareSendTransaction(from, to, amount, candidates, client)
    }

    // pass 4: randomly permute utxos until sum is greater than amount or cap is reached
    for i := 0; i < 1000; i++ {
        lIdx := rand.Intn(len(txs))
        rIdx := lIdx

        for lIdx == rIdx {
            rIdx = rand.Intn(len(txs))
        }

        ltx := txs[lIdx]
        rtx := txs[rIdx]

        total := big.NewInt(0).Add(ltx.OutputFor(&from).Amount, rtx.OutputFor(&from).Amount)

        if total.Cmp(amount) == 1 {
            candidates = append(candidates, ltx)
            candidates = append(candidates, rtx)
            return PrepareSendTransaction(from, to, amount, candidates, client)
        }
    }

    return nil, errors.New("no suitable UTXOs found")
}

func PrepareSendTransaction(from, to common.Address, amount *big.Int, utxoTxs []Transaction, client plasma_common.Client) (*Transaction, error) {
    var input1 *Input
    var output1 *Output

    if len(utxoTxs) == 1 {
        input1 = ZeroInput()

        utxo := utxoTxs[0].OutputFor(&from)

        if utxo == nil {
            return nil, errors.New("expected a UTXO")
        }

        totalAmount := utxo.Amount

        output1 = &Output{
            NewOwner: from,
            Amount:   big.NewInt(0).Sub(totalAmount, amount),
        }
    } else {
        input1 = &Input{
            BlkNum: utxoTxs[1].BlkNum,
            TxIdx:  utxoTxs[1].TxIdx,
            OutIdx: utxoTxs[1].OutputIndexFor(&from),
        }

        totalAmount := big.NewInt(0)
        totalAmount = totalAmount.Add(utxoTxs[0].OutputFor(&from).Amount, utxoTxs[1].OutputFor(&from).Amount)

        output1 = &Output{
            NewOwner: from,
            Amount:   big.NewInt(0).Sub(totalAmount, amount),
        }
    }

    tx := Transaction{
        Input0: &Input{
            BlkNum: utxoTxs[0].BlkNum,
            TxIdx:  utxoTxs[0].TxIdx,
            OutIdx: utxoTxs[0].OutputIndexFor(&from),
        },
        Input1: input1,
        Output0: &Output{
            NewOwner: to,
            Amount:   amount,
        },
        Output1: output1,
        Fee:     big.NewInt(0),
    }
    var err error
    tx.Sig0, err = client.SignData(&from, tx.SignatureHash())
    if err != nil {
        return nil, err
    }

    tx.Sig1, err = client.SignData(&from, tx.SignatureHash())
    if err != nil {
        return nil, err
    }
    return &tx, nil
}
