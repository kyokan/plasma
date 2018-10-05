package cmd

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/cli"
	"github.com/spf13/cobra"
	"math/big"
)

const FlagTo = "to"

var sendCmd = &cobra.Command{
	Use: "send",
	Short: "sends funds",
	RunE: func(cmd *cobra.Command, args []string) error {
		privateKey, err := ParsePrivateKey()
		if err != nil {
			return err
		}
		rpcHost, err := cmd.Flags().GetString(FlagRPCPort)
		if err != nil {
			return err
		}
		amountStr, err := cmd.Flags().GetString(FlagAmount)
		if err != nil {
			return err
		}
		amount, ok := new(big.Int).SetString(amountStr, 10)
		if !ok {
			return errors.New("invalid amount")
		}
		fromStr, err := cmd.Flags().GetString(FlagAddress)
		if err != nil {
			return err
		}
		from := common.HexToAddress(fromStr)
		toStr, err := cmd.Flags().GetString(FlagTo)
		if err != nil {
			return err
		}
		to := common.HexToAddress(toStr)
		return cli.Send(privateKey, rpcHost, from, to, amount)
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
	sendCmd.Flags().String(FlagRPCPort, "", "hostname and port of the root node")
	sendCmd.Flags().String(FlagAddress, "", "address to send funds from")
	sendCmd.Flags().String(FlagTo, "", "address to send funds to")
	sendCmd.Flags().String(FlagAmount, "", "amount to send")
	sendCmd.MarkFlagRequired(FlagRPCPort)
	sendCmd.MarkFlagRequired(FlagAddress)
	sendCmd.MarkFlagRequired(FlagTo)
	sendCmd.MarkFlagRequired(FlagAmount)
}