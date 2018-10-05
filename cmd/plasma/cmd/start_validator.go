package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/kyokan/plasma/validator"
)

const FlagRootHost = "root-host"

var startValidatorCommand = &cobra.Command{
	Use:   "start-validator",
	Short: "starts running a Plasma validator node",
	RunE: func(cmd *cobra.Command, args []string) error {
		privateKey, err := ParsePrivateKey()
		if err != nil {
			return err
		}

		return validator.Start(NewGlobalConfig(), privateKey, viper.GetString(FlagRootHost))
	},
}

func init() {
	rootCmd.AddCommand(startValidatorCommand)
	startValidatorCommand.Flags().Uint(FlagRPCPort, 8643, "port for the RPC server to listen on")
	startValidatorCommand.Flags().String(FlagRootHost, "", "hostname and port of the root node")
	startValidatorCommand.MarkFlagRequired(FlagRootHost)
}
