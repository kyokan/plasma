package cmd

import (
	"github.com/kyokan/plasma/cli"
	"github.com/spf13/cobra"
)

const (
	FlagTxIndex  = "txindex"
	FlagOIndex   = "oindex"
)

var exitCmd = &cobra.Command{
	Use:   "exit",
	Short: "performs an exit",
	RunE: func(cmd *cobra.Command, args []string) error {
		privateKey, err := ParsePrivateKey()
		if err != nil {
			return err
		}

		rootHost, err := cmd.Flags().GetString(FlagRootHost)
		if err != nil {
			return err
		}
		blockNum, err := cmd.Flags().GetUint64(FlagBlockNum)
		if err != nil {
			return err
		}
		txIndex, err := cmd.Flags().GetUint32(FlagTxIndex)
		if err != nil {
			return err
		}
		oIndex, err := cmd.Flags().GetUint8(FlagOIndex)
		if err != nil {
			return err
		}
		return cli.Exit(NewGlobalConfig(), privateKey, rootHost, blockNum, txIndex, oIndex)
	},
}

func init() {
	rootCmd.AddCommand(exitCmd)
	exitCmd.Flags().String(FlagRootHost, "", "hostname and port of the root node")
	exitCmd.Flags().String(FlagBlockNum, "", "block number to exit")
	exitCmd.Flags().Uint(FlagTxIndex, 0, "transaction to exit")
	exitCmd.Flags().Uint(FlagOIndex, 0, "output to exit")
	exitCmd.MarkFlagRequired(FlagRootHost)
	exitCmd.MarkFlagRequired(FlagBlockNum)
	exitCmd.MarkFlagRequired(FlagTxIndex)
	exitCmd.MarkFlagRequired(FlagOIndex)
}
