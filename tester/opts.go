package tester

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func CreateAuth(privateKeyECDSA *ecdsa.PrivateKey) *bind.TransactOpts {
	auth := bind.NewKeyedTransactor(privateKeyECDSA)
	auth.GasPrice = new(big.Int).SetUint64(1)
	auth.GasLimit = uint64(4712388)
	return auth
}

func CreateCallOpts(address string) *bind.CallOpts {
	return &bind.CallOpts{
		From:    common.HexToAddress(address),
		Context: context.Background(),
	}
}
