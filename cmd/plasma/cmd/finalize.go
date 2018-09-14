package cmd

import "github.com/spf13/cobra"

var finalizeCmd = &cobra.Command{
	Use: "finalize",
	Short: "finalizes a transaction",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(finalizeCmd)
}