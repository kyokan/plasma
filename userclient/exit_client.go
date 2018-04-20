package userclient

import (
	"fmt"

	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/util"
	"github.com/urfave/cli"
)

// TODO: move to client with sub args for deposit args.
func StartExit(c *cli.Context) {
	contractAddress := c.GlobalString("contract-addr")
	nodeURL := c.GlobalString("node-url")
	keystoreDir := c.GlobalString("keystore-dir")
	keystoreFile := c.GlobalString("keystore-file")
	userAddress := c.GlobalString("user-address")
	privateKey := c.GlobalString("private-key")
	signPassphrase := c.GlobalString("sign-passphrase")
	useGeth := c.GlobalBool("use-geth")

	// Used for starting exit.
	rootPort := c.Int("root-port")
	blocknum := c.Int("blocknum")
	txindex := c.Int("txindex")
	oindex := c.Int("oindex")

	fmt.Printf("Exit starting for blocknum: %d, txindex: %d, oindex: %d", blocknum, txindex, oindex)

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

	rootUrl := fmt.Sprintf("http://localhost:%d/rpc", rootPort)

	res := GetBlock(rootUrl, uint64(blocknum))

	plasma.StartExit(
		res.Block,
		res.Transactions,
		util.NewInt(blocknum),
		util.NewInt(txindex),
		util.NewInt(oindex),
	)
}
