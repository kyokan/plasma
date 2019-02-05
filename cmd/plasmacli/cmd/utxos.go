package cmd

import (
	"github.com/spf13/cobra"
	"context"
	"time"
	"github.com/kyokan/plasma/rpc/pb"
	"github.com/kyokan/plasma/rpc"
	"github.com/kyokan/plasma/util"
)

type utxoCmdOutput struct {
	BlockNumber      uint64 `json:"blockNumber"`
	TransactionIndex uint32 `json:"transactionIndex"`
	OutputIndex      uint8  `json:"outputIndex"`
	Amount           string `json:"amount"`
	DepositNonce     string `json:"depositNonce"`
}

var utxosCmd = &cobra.Command{
	Use:   "utxos [addr]",
	Short: "Returns the UTXOs for a given address",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := AddrOrPrivateKeyAddr(cmd, args, 0)
		if err != nil {
			return err
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

		out := make([]utxoCmdOutput, len(res.ConfirmedTransactions), len(res.ConfirmedTransactions))

		for i, conf := range res.ConfirmedTransactions {
			deser := rpc.DeserializeConfirmedTx(conf)
			tx := deser.Transaction

			out[i] = utxoCmdOutput{
				BlockNumber:      util.Big2Uint64(tx.BlkNum),
				TransactionIndex: util.Big2Uint32(tx.TxIdx),
				OutputIndex:      tx.OutputIndexFor(&addr),
				Amount:           tx.OutputFor(&addr).Denom.Text(10),
				DepositNonce:     tx.OutputFor(&addr).DepositNonce.Text(10),
			}
		}

		return PrintJSON(out)
	},
}

func init() {
	rootCmd.AddCommand(utxosCmd)
}
