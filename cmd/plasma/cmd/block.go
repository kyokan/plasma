package cmd

import (
	"github.com/kyokan/plasma/cli"
	"github.com/spf13/cobra"
)

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "outputs transactions in a block",
	RunE: func(cmd *cobra.Command, args []string) error {
		rootHost, err := cmd.Flags().GetString(FlagRootHost)
		if err != nil {
			return err
		}
		blockNum, err := cmd.Flags().GetUint64(FlagBlockNum)
		if err != nil {
			return err
		}

		return cli.Block(rootHost, blockNum)
	},
}

func init() {
	rootCmd.AddCommand(blockCmd)
	blockCmd.Flags().String(FlagRootHost, "", "hostname and port of the root node")
	blockCmd.Flags().String(FlagBlockNum, "", "the block number to show")
	blockCmd.MarkFlagRequired(FlagRootHost)
	blockCmd.MarkFlagRequired(FlagBlockNum)
}
