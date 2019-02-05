package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
	"os"
)

var rootCmd = &cobra.Command{
	Use: "plasmacli",
	Short: "Interacts with a running plasmad instance.",
}

func init() {
	rootCmd.PersistentFlags().StringP(FlagPrivateKeyPath, "p", "~/.plasma/key", "Path to your private key.")
	rootCmd.PersistentFlags().StringP(FlagNodeURL, "u", "localhost:6545", "URL to a running plasmad instance.")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
	    fmt.Println(err)
	    os.Exit(1)
	}
}