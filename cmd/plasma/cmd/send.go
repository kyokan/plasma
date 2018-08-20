package cmd

import "github.com/spf13/cobra"

const FlagTo = "to"

var sendCmd = &cobra.Command{
	Use: "send",
	Short: "sends funds",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
	sendCmd.Flags().String(FlagRPCPort, "", "hostname and port of the root node")
	sendCmd.Flags().String(FlagTo, "", "address to send funds to")
	sendCmd.Flags().String(FlagAmount, "", "amount to send")
	sendCmd.MarkFlagRequired(FlagRPCPort)
	sendCmd.MarkFlagRequired(FlagTo)
	sendCmd.MarkFlagRequired(FlagAmount)
}