package cmd

import (
	"github.com/spf13/cobra"
	"github.com/kyokan/plasma/pkg/eth"
	"math/big"
	"github.com/pkg/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/pkg/chain"
	"bytes"
)

type depositCmdOutput struct {
	TransactionHash string `json:"transactionHash"`
	ContractAddress string `json:"contractAddress"`
	Amount          string `json:"amount"`
	DepositNonce    string `json:"depositNonce"`
}

type DepositReceipt struct {
	TransactionHash common.Hash
	ContractAddress common.Address
	Amount          *big.Int
	DepositNonce    *big.Int
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

		receipt, err := Deposit(client, amount)

		return PrintJSON(&depositCmdOutput{
			TransactionHash: receipt.TransactionHash.Hex(),
			ContractAddress: receipt.ContractAddress.Hex(),
			Amount:          amount.Text(10),
			DepositNonce:    receipt.DepositNonce.Text(10),
		})
	},
}

func Deposit(client eth.Client, amount *big.Int) (*DepositReceipt, error) {
	receipt, err := client.Deposit(amount)
	if err != nil {
		return nil, err
	}

	logData := receipt.Logs[0].Data
	dataReader := bytes.NewReader(logData)
	var depositNonce chain.UInt256
	_, err = dataReader.ReadAt(depositNonce[:], 64)
	if err != nil {
		return nil, err
	}

	return &DepositReceipt{
		TransactionHash: receipt.TxHash,
		ContractAddress: receipt.ContractAddress,
		Amount:          amount,
		DepositNonce:    depositNonce.ToBig(),
	}, nil
}

func init() {
	rootCmd.AddCommand(depositCmd)
	depositCmd.Flags().StringP(FlagEthereumNodeUrl, "e", "http://localhost:8545", "URL to a running Ethereum node.")
}
