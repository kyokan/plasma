package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/kyokan/plasma/internal/root"
)

var startRootCmd = &cobra.Command{
	Use:   "start-root",
	Short: "starts running a Plasma root node",
	RunE: func(cmd *cobra.Command, args []string) error {
		privateKey, err := ParsePrivateKey()
		if err != nil {
			return err
		}

		return root.Start(NewGlobalConfig(), privateKey)
	},
}

func init() {
	rootCmd.AddCommand(startRootCmd)
	startRootCmd.Flags().Uint(FlagRPCPort, 6545, "port for the RPC server to listen on")
	startRootCmd.Flags().Uint(FlagRESTPort, 6546, "port for the RPC server to listen on")
	viper.BindPFlag(FlagRPCPort, startRootCmd.Flags().Lookup(FlagRPCPort))
	viper.BindPFlag(FlagRESTPort, startRootCmd.Flags().Lookup(FlagRESTPort))
}
