package userclient

import (
	"log"

	"github.com/kyokan/plasma/db"
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
	dburl := c.GlobalString("db")
	useGeth := c.GlobalBool("use-geth")

	// Used for starting exit.
	blocknum := c.Int("blocknum")
	txindex := c.Int("txindex")
	oindex := c.Int("oindex")

	db, level, err := db.CreateLevelDatabase(dburl)

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

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

	block, err := level.BlockDao.BlockAtHeight(uint64(blocknum))

	if err != nil {
		panic(err)
	}

	txs, err := level.TxDao.FindByBlockNum(uint64(blocknum))

	if err != nil {
		panic(err)
	}

	plasma.StartExit(
		block,
		txs,
		util.NewInt(blocknum),
		util.NewInt(txindex),
		util.NewInt(oindex),
	)
}
