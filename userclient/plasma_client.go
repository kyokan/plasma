package userclient

import (
	"fmt"
	"log"
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

func StartExit(c *cli.Context) {
	plasma := eth.CreatePlasmaClientCLI(c)

	rootUrl := fmt.Sprintf("http://localhost:%d/rpc", c.Int("root-port"))
	blocknum := c.Int("blocknum")
	txindex := c.Int("txindex")
	oindex := c.Int("oindex")

	fmt.Printf("Exit starting for blocknum: %d, txindex: %d, oindex: %d\n", blocknum, txindex, oindex)

	rootClient := NewRootClient(rootUrl)
	res := rootClient.GetBlock(uint64(blocknum))

	if res == nil {
		log.Fatalln("Block does not exist!")
	}

	plasma.StartExit(
		res.Block,
		res.Transactions,
		util.NewInt(blocknum),
		util.NewInt(txindex),
		util.NewInt(oindex),
	)
}

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
		log.Fatalf("Failed to get current child block: %v", err)
	}

	fmt.Printf("Last child block: %v\n", curr)
}

// TODO: Use same code as transaction sink.
func createDepositTx(userAddress string, value uint64) chain.Transaction {
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
