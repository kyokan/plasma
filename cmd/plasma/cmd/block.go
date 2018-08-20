package cmd

import (
	"github.com/spf13/cobra"
	"github.com/kyokan/plasma/cli"
	"math/big"
	"github.com/pkg/errors"
)

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "outputs transactions in a block",
	RunE: func(cmd *cobra.Command, args []string) error {
		rootHost, err := cmd.Flags().GetString(FlagRootHost)
		if err != nil {
			return err
		}
		blockNum, err := cmd.Flags().GetString(FlagBlockNum)
		if err != nil {
			return err
		}

		blockNumBig, ok := new(big.Int).SetString(blockNum, 10)
		if !ok {
			return errors.New("invalid block number")
		}

		return cli.Block(rootHost, blockNumBig)
	},
}

func init() {
	rootCmd.AddCommand(blockCmd)
	blockCmd.Flags().String(FlagRootHost, "", "hostname and port of the root node")
	blockCmd.Flags().String(FlagBlockNum, "", "the block number to show")
	blockCmd.MarkFlagRequired(FlagRootHost)
	blockCmd.MarkFlagRequired(FlagBlockNum)
}
