package userclient

import (
	"fmt"
	"math/big"

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
	amount := c.Int("amount")

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
}

// TODO: move this to utils
func createDepositTx(userAddress string, value int) chain.Transaction {
	return createTransaction(
		chain.ZeroInput(),
		&chain.Output{
			NewOwner: common.HexToAddress(userAddress),
			Amount:   util.NewInt(value),
		},
	)
}

func createTransaction(
	input0 *chain.Input,
	output0 *chain.Output,
) chain.Transaction {
	return chain.Transaction{
		Input0:  input0,
		Input1:  chain.ZeroInput(),
		Sig0:    []byte{},
		Sig1:    []byte{},
		Output0: output0,
		Output1: chain.ZeroOutput(),
		Fee:     new(big.Int),
		BlkNum:  uint64(0),
		TxIdx:   0,
	}
}
