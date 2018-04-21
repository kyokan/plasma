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

// TODO: move to client with sub args for deposit args.
func Deposit(c *cli.Context) {
	contractAddress := c.GlobalString("contract-addr")
	nodeURL := c.GlobalString("node-url")
	keystoreDir := c.GlobalString("keystore-dir")
	keystoreFile := c.GlobalString("keystore-file")
	userAddress := c.GlobalString("user-address")
	privateKey := c.GlobalString("private-key")
	signPassphrase := c.GlobalString("sign-passphrase")
	useGeth := c.GlobalBool("use-geth")

	// Used for starting exit.
	amount := uint64(c.Int("amount"))

	fmt.Printf("Deposit starting for amount: %d\n", amount)

	privateKeyECDSA := util.CreatePrivateKeyECDSA(
		userAddress,
		privateKey,
		keystoreDir,
		keystoreFile,
		signPassphrase,
	)

	plasma := eth.CreatePlasmaClient(
		nodeURL,
		contractAddress,
		userAddress,
		privateKeyECDSA,
		useGeth,
	)

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
