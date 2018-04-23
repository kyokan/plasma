package userclient

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/util"
	"github.com/urfave/cli"
)

func Finalize(c *cli.Context) {
	plasma := eth.CreatePlasmaClientCLI(c)
	plasma.Finalize()
}

// TODO: move to client with sub args for deposit args.
func StartExit(c *cli.Context) {
	plasma := eth.CreatePlasmaClientCLI(c)

	// Used for starting exit.
	rootPort := c.Int("root-port")
	blocknum := c.Int("blocknum")
	txindex := c.Int("txindex")
	oindex := c.Int("oindex")

	fmt.Printf("Exit starting for blocknum: %d, txindex: %d, oindex: %d\n", blocknum, txindex, oindex)

	rootUrl := fmt.Sprintf("http://localhost:%d/rpc", rootPort)

	// TODO: is the hash i'm exiting with the wrong one?
	res := GetBlock(rootUrl, uint64(blocknum))

	if res == nil {
		panic("Block does not exist!")
	}

	plasma.StartExit(
		res.Block,
		res.Transactions,
		util.NewInt(blocknum),
		util.NewInt(txindex),
		util.NewInt(oindex),
	)
}

// TODO: move to client with sub args for deposit args.
func Deposit(c *cli.Context) {
	plasma := eth.CreatePlasmaClientCLI(c)

	userAddress := c.GlobalString("user-address")
	amount := uint64(c.Int("amount"))

	fmt.Printf("Deposit starting for amount: %d\n", amount)

	t := createDepositTx(userAddress, amount)

	plasma.Deposit(amount, &t)

	time.Sleep(3 * time.Second)

	curr, err := plasma.CurrentChildBlock()

	if err != nil {
		panic(err)
	}

	fmt.Println("**** deposit last child block")
	fmt.Println(curr.Uint64())
}

// TODO: use same code as transaction sink.
func createDepositTx(userAddress string, value uint64) chain.Transaction {
	fmt.Println("***** createDepositTx")
	fmt.Println(value)
	return chain.Transaction{
		Input0: chain.ZeroInput(),
		Input1: chain.ZeroInput(),
		Output0: &chain.Output{
			NewOwner: common.HexToAddress(userAddress),
			Amount:   util.NewUint64(value),
		},
		Output1: chain.ZeroOutput(),
		Fee:     big.NewInt(0),
	}
}
