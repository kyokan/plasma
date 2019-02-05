package cmd

import (
	"github.com/spf13/cobra"
	"github.com/kyokan/plasma/eth"
	"math/big"
	"github.com/pkg/errors"
)

type depositCmdOutput struct {
	TransactionHash string `json:"transactionHash"`
	ContractAddress string `json:"contractAddress"`
	Amount          string `json:"amount"`
}

var depositCmd = &cobra.Command{
	Use:   "deposit addr amount",
	Short: "Deposits funds into the Plasma smart contract",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		amount, valid := new(big.Int).SetString(args[1], 10)
		if !valid {
			return errors.New("invalid amount")
		}

		privKey, err := ParsePrivateKey(cmd)
		if err != nil {
			return err
		}

		client, err := eth.NewClient(cmd.Flag(FlagEthereumNodeUrl).Value.String(), args[0], privKey)
		if err != nil {
			return err
		}

		receipt, err := client.Deposit(amount)
		if err != nil {
			return err
		}

		return PrintJSON(&depositCmdOutput{
			TransactionHash: receipt.TxHash.Hex(),
			ContractAddress: args[0],
			Amount:          amount.Text(10),
		})
	},
}

func init() {
	rootCmd.AddCommand(depositCmd)
	depositCmd.Flags().StringP(FlagEthereumNodeUrl, "e", "http://localhost:8545", "URL to a running Ethereum node.")
}
