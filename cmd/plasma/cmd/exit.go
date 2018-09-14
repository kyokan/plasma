package cmd

import (
	"github.com/spf13/cobra"
			"github.com/kyokan/plasma/cli"
	"math/big"
	"github.com/pkg/errors"
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
		blockNumStr, err := cmd.Flags().GetString(FlagBlockNum)
		if err != nil {
			return err
		}
		blockNumBig, ok := new(big.Int).SetString(blockNumStr, 10)
		if !ok {
			return errors.New("invalid block number")
		}

		txIndex, err := cmd.Flags().GetUint(FlagTxIndex)
		if err != nil {
			return err
		}
		oIndex, err := cmd.Flags().GetUint(FlagOIndex)
		if err != nil {
			return err
		}
		return cli.Exit(NewGlobalConfig(), privateKey, rootHost, blockNumBig, txIndex, oIndex)
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
