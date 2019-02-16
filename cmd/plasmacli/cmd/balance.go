package cmd

import (
	"github.com/spf13/cobra"
			"context"
	"time"
	"github.com/kyokan/plasma/rpc/pb"
	"github.com/kyokan/plasma/rpc"
	"errors"
)

type balanceCmdOutput struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
}

var balanceCmd = &cobra.Command{
	Use: "balance [address]",
	Short: "Returns balance for an account",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := AddrOrPrivateKeyAddr(cmd, args, 0)
		if err != nil {
			return err
		}

		url := cmd.Flag(FlagNodeURL).Value.String()
		if url == "" {
			return errors.New("no node url set")
		}
		client, conn, err := CreateRootClient(url)
		if err != nil {
			return err
		}
		defer conn.Close()

		ctx, _ := context.WithTimeout(context.Background(), time.Second * 5)
		res, err := client.GetBalance(ctx, &pb.GetBalanceRequest{
			Address: addr.Bytes(),
		})
		if err != nil {
			return err
		}

		balance := rpc.DeserializeBig(res.Balance)
		out := &balanceCmdOutput{
			Address: addr.Hex(),
			Balance: balance.Text(10),
		}
		return PrintJSON(out)
	},
}

func init() {
	rootCmd.AddCommand(balanceCmd)
}