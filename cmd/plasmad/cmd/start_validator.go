package cmd

import (
	"github.com/spf13/cobra"
	"github.com/kyokan/plasma/internal/validator"
	"github.com/spf13/viper"
)

const FlagRootURL = "root-url"

var startValidatorCmd = &cobra.Command{
	Use:   "start-validator",
	Short: "starts running a Plasma validator node",
	RunE: func(cmd *cobra.Command, args []string) error {
		privateKey, err := ParsePrivateKey()
		if err != nil {
			return err
		}

		return validator.Start(NewGlobalConfig(), viper.GetString(FlagRootURL), privateKey)
	},
}

func init() {
	rootCmd.AddCommand(startValidatorCmd)
	startValidatorCmd.Flags().String(FlagRootURL, "localhost:6545", "URL belonging to the root node")
	startValidatorCmd.Flags().Uint(FlagRPCPort, 6545, "port for the RPC server to listen on")
	viper.BindPFlag(FlagRootURL, startValidatorCmd.Flags().Lookup(FlagRootURL))
	viper.BindPFlag(FlagRPCPort, startValidatorCmd.Flags().Lookup(FlagRPCPort))
}
