package types

import (
	"github.com/kyokan/plasma/chain"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type TransactionRequest struct {
	chain.Transaction
	From     common.Address
	To       common.Address
	Amount   *big.Int
	Response *TransactionResponse
}

type TransactionResponse struct {
	Error       error
	Transaction *chain.Transaction
}

