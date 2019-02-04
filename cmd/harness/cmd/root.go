package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "harness",
	Short: "Manages a local development test harness.",
}

func init() {
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
	    fmt.Println(err)
	    os.Exit(1)
	}
}