package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"crypto/ecdsa"
	"math/big"
	"context"
)

func CreateCallOpts(address common.Address) *bind.CallOpts {
	return &bind.CallOpts{
		From:    address,
		Context: context.Background(),
	}
}

func CreateKeyedTransactor(privateKey *ecdsa.PrivateKey) *bind.TransactOpts {
	auth := bind.NewKeyedTransactor(privateKey)
	auth.GasPrice = new(big.Int).SetUint64(10 * 1000000000)
	auth.GasLimit = uint64(4712388)
	return auth
}