package eth

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/kyokan/plasma/util"
	"time"
	"errors"
	"github.com/ethereum/go-ethereum/ethclient"
	"context"
)

type TxGenerator = func () (*types.Transaction, error)

func ContractCall(client *ethclient.Client, generator TxGenerator) (*types.Receipt, error) {
	tx, err := generator()
	if err != nil {
		return nil, err
	}

	rawReceipt, err := util.WithRetries(func() (interface{}, error) {
		return client.TransactionReceipt(context.Background(), tx.Hash())
	}, 24, 5 * time.Second)
	if err != nil {
		return nil, err
	}
	if rawReceipt == nil {
		return nil, errors.New("receipt is nil")
	}

	receipt := rawReceipt.(*types.Receipt)
	if receipt.Status == 0 {
		return nil, errors.New("transaction reverted")
	}
	if receipt.GasUsed == tx.Gas() {
		return nil, errors.New("transaction ran out of gas (usually implies contract bug)")
	}

	return receipt, nil
}