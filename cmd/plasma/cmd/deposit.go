package cmd

import (
	"github.com/spf13/cobra"
	"github.com/kyokan/plasma/cli"
	"math/big"
	"github.com/pkg/errors"
)

var depositCmd = &cobra.Command{
	Use: "deposit",
	Short: "performs a deposit",
	RunE: func(cmd *cobra.Command, args []string) error {
		privateKey, err := ParsePrivateKey()
		if err != nil {
			return err
		}

		amountStr, err := cmd.Flags().GetString(FlagAmount)
		if err != nil {
			return err
		}
		amount, ok := new(big.Int).SetString(amountStr, 10)
		if !ok {
			return errors.New("invalid amount")
		}

		return cli.Deposit(NewGlobalConfig(), privateKey, amount)
	},
}

func init() {
	rootCmd.AddCommand(depositCmd)
	depositCmd.Flags().String(FlagAmount, "", "the amount to deposit")
	depositCmd.MarkFlagRequired(FlagAmount)
}