package common

import (
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
)

type Client interface {
    GetBalance(addr common.Address) (*big.Int, error)
    SignData(addr *common.Address, data []byte) ([]byte, error)
    NewGethTransactor(keyAddr common.Address) *bind.TransactOpts
}
