package cmd

import (
	"github.com/spf13/cobra"
	"time"
	"github.com/kyokan/plasma/rpc/pb"
	"github.com/ethereum/go-ethereum/crypto"
	"context"
	"github.com/kyokan/plasma/rpc"
	"errors"
	"github.com/kyokan/plasma/chain"
	"sort"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
		"github.com/kyokan/plasma/util"
	"github.com/kyokan/plasma/eth"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type sendCmdOutput struct {
	Value string `json:"value"`
	To    string `json:"to"`
	ConfirmationSignatures []string `json:"confirmationSignatures"`
}

var sendCmd = &cobra.Command{
	Use:   "send to value",
	Short: "Sends funds",
	RunE: func(cmd *cobra.Command, args []string) error {
		privKey, err := ParsePrivateKey(cmd)
		if err != nil {
			return err
		}
		addr := crypto.PubkeyToAddress(privKey.PublicKey)
		to := common.HexToAddress(args[0])
		value, ok := new(big.Int).SetString(args[1], 10)
		if !ok {
			return errors.New("invalid send value")
		}

		client, conn, err := CreateRootClient(cmd)
		if err != nil {
			return err
		}
		defer conn.Close()

		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
		res, err := client.GetOutputs(ctx, &pb.GetOutputsRequest{
			Address:   addr.Bytes(),
			Spendable: true,
		})
		if err != nil {
			return err
		}
		utxos := rpc.DeserializeConfirmedTxs(res.ConfirmedTransactions)
		if len(utxos) == 0 {
			return errors.New("no spendable outputs")
		}
		selectedUtxos, err := selectUTXOs(utxos, addr, value)
		if err != nil {
			return err
		}

		total := big.NewInt(0)
		tx := chain.ZeroTransaction()
		for i, utxo := range selectedUtxos {
			var input *chain.Input
			if i == 0 {
				input = tx.Input0
			} else if i == 1 {
				input = tx.Input1
			} else {
				panic("too many inputs!")
			}

			input.BlkNum = util.Big2Uint64(utxo.BlkNum)
			input.TxIdx = util.Big2Uint32(utxo.TxIdx)
			input.OutIdx = utxo.OutputIndexFor(&addr)
			input.Owner = addr
			sig, err := eth.Sign(privKey, input.SignatureHash())
			if err != nil {
				return err
			}

			if i == 0 {
				tx.Sig0 = sig
			} else {
				tx.Sig1 = sig
			}

			total = total.Add(total, utxo.OutputFor(&addr).Denom)
		}

		tx.Output0.Denom = value
		tx.Output0.Owner = to

		if total.Cmp(value) > 0 {
			totalClone := new(big.Int).Set(total)
			tx.Output1.Denom = totalClone.Sub(totalClone, value)
			tx.Output1.Owner = addr
		}

		confirmSig, err := eth.Sign(privKey, tx.SignatureHash())
		if err != nil {
			return err
		}

		confirmed := &chain.ConfirmedTransaction{
			Transaction: *tx,
			Signatures: [2]chain.Signature{
				confirmSig,
				confirmSig,
			},
		}

		ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
		_, err = client.Send(ctx, &pb.SendRequest{
			Confirmed: rpc.SerializeConfirmedTx(confirmed),
		})
		if err != nil {
			return err
		}

		out := &sendCmdOutput{
			Value: value.Text(10),
			To: to.Hex(),
			ConfirmationSignatures: []string {
				hexutil.Encode(confirmSig[:]),
				hexutil.Encode(confirmSig[:]),
			},
		}

		return PrintJSON(out)
	},
}

func selectUTXOs(confirmedTxs []chain.ConfirmedTransaction, addr common.Address, total *big.Int) ([]chain.Transaction, error) {
	sort.Slice(confirmedTxs, func(i, j int) bool {
		a := confirmedTxs[i].Transaction.OutputFor(&addr).Denom
		b := confirmedTxs[j].Transaction.OutputFor(&addr).Denom
		return a.Cmp(b) > 0
	})

	first := confirmedTxs[0].Transaction

	if first.OutputFor(&addr).Denom.Cmp(total) >= 0 {
		return []chain.Transaction{first}, nil
	}

	for i := len(confirmedTxs) - 1; i >= 0; i-- {
		tx := confirmedTxs[i].Transaction
		sum := big.NewInt(0)
		sum = sum.Add(sum, first.OutputFor(&addr).Denom)
		sum = sum.Add(sum, tx.OutputFor(&addr).Denom)
		if sum.Cmp(total) >= 0 {
			return []chain.Transaction{
				first,
				tx,
			}, nil
		}
	}

	return nil, errors.New("no suitable UTXOs found")
}

func init() {
	rootCmd.AddCommand(sendCmd)
}
