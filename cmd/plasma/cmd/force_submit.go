package cmd

import "github.com/spf13/cobra"

const (
	FlagMerkleRoot = "merkle-root"
	FlagPrevHash       = "prev-hash"
)

var forceSubmitCmd = &cobra.Command{
	Use:   "force-submit",
	Short: "force submits a block to the contract",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(forceSubmitCmd)
	forceSubmitCmd.Flags().String(FlagMerkleRoot, "", "the merkle root of the block to submit")
	forceSubmitCmd.Flags().String(FlagPrevHash, "", "the hash of the previous block")
	forceSubmitCmd.Flags().String(FlagBlockNum, "", "the block number")
	forceSubmitCmd.MarkFlagRequired(FlagMerkleRoot)
	forceSubmitCmd.MarkFlagRequired(FlagPrevHash)
	forceSubmitCmd.MarkFlagRequired(FlagBlockNum)
}
