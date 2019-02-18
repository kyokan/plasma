package chain

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"sync"
)

var zero, one *big.Int
var exitOutput *Output
var once sync.Once

func initialize() {
	zero = big.NewInt(0)
	one  = big.NewInt(1)
	exitOutput = ZeroOutput()
	exitOutput.Owner = common.HexToAddress("0xDeadBeefCafe")
}

func Zero() *big.Int {
	once.Do(initialize)
	return zero
}

func One() *big.Int {
	once.Do(initialize)
	return one
}

func ExitOutput() *Output {
	once.Do(initialize)
	return exitOutput
}

