package cmd

import (
	"os"

	"github.com/kyokan/plasma/db"
	"github.com/spf13/cobra"
	"fmt"
	"github.com/spf13/viper"
)

const (
	FlagConfig       = "config"
	FlagDB           = "db"
	FlagNodeURL      = "node-url"
	FlagContractAddr = "contract-addr"
	FlagPrivateKey   = "private-key"
	FlagRPCPort      = "rpc-port"
	FlagAddress      = "address"
	FlagBlockNum     = "blocknum"
	FlagAmount       = "amount"
)

var boundFlags = []string{
	FlagDB,
	FlagNodeURL,
	FlagContractAddr,
	FlagPrivateKey,
}

var configFile string

var rootCmd = &cobra.Command{
	Use:   "plasma",
	Short: "An implementation of the Minimum Viable Plasma spec.",
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&configFile, FlagConfig, "", "filepath to Plasma's configuration file")
	rootCmd.PersistentFlags().String(FlagDB, db.DefaultLocation(), "filepath to Plasma's database")
	rootCmd.PersistentFlags().String(FlagNodeURL, "", "full URL to a running Ethereum node")
	rootCmd.PersistentFlags().String(FlagContractAddr, "", "address of the Plasma contract")
	rootCmd.PersistentFlags().String(FlagPrivateKey, "", "node operator's private key")
	for _, flag := range boundFlags {
		rootCmd.MarkFlagRequired(flag)
		viper.BindPFlag(flag, rootCmd.PersistentFlags().Lookup(flag))
	}
}

func initConfig() {
	if configFile == "" {
		return
	}

	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Invalid config file:", err)
		os.Exit(1)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
