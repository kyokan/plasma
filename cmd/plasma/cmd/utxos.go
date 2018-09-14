package cmd

import (
	"github.com/spf13/cobra"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/cli"
)

var utxosCmd = &cobra.Command{
	Use: "utxos",
	Short: "prints UTXOs for a given address",
	RunE: func(cmd *cobra.Command, args []string) error {
		rootHost, err := cmd.Flags().GetString(FlagRootHost)
		if err != nil {
			return err
		}
		addrStr, err := cmd.Flags().GetString(FlagAddress)
		if err != nil {
			return err
		}

		addr := common.HexToAddress(addrStr)
		return cli.UTXOs(rootHost, addr)
	},
}

func init() {
	rootCmd.AddCommand(utxosCmd)
	utxosCmd.Flags().String(FlagAddress, "", "the address to print UTXOs for")
	utxosCmd.Flags().String(FlagRootHost, "", "the hostname and port of the root node")
	utxosCmd.MarkFlagRequired(FlagAddress)
	utxosCmd.MarkFlagRequired(FlagRootHost)
}