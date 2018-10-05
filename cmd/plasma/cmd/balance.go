package cmd

import (
	"github.com/spf13/cobra"
	"github.com/kyokan/plasma/cli"
	"github.com/ethereum/go-ethereum/common"
)

var balanceCmd = &cobra.Command{
	Use: "balance",
	Short: "shows an address's balance",
	RunE: func(cmd *cobra.Command, args []string) error {
		rootHost, err := cmd.Flags().GetString(FlagRootHost)
		if err != nil {
			return err
		}
		addressStr, err := cmd.Flags().GetString(FlagAddress)
		if err != nil {
			return err
		}
		address := common.HexToAddress(addressStr)
		return cli.Balance(rootHost, address)
	},
}

func init() {
	balanceCmd.Flags().String(FlagAddress, "", "the address to show balances for")
	balanceCmd.Flags().String(FlagRootHost, "", "the hostname and port of the root node")
	balanceCmd.MarkFlagRequired(FlagAddress)
	balanceCmd.MarkFlagRequired(FlagRootHost)
	rootCmd.AddCommand(balanceCmd)
}