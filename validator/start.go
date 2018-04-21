package validator

import (
	"fmt"
	"log"
	"path"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/util"
	"github.com/urfave/cli"
)

func Start(c *cli.Context) {
	fmt.Println("Validator Starting")

	contractAddress := c.GlobalString("contract-addr")
	nodeURL := c.GlobalString("node-url")
	keystoreDir := c.GlobalString("keystore-dir")
	keystoreFile := c.GlobalString("keystore-file")
	userAddress := c.GlobalString("user-address")
	privateKey := c.GlobalString("private-key")
	signPassphrase := c.GlobalString("sign-passphrase")
	dburl := c.GlobalString("db")
	useGeth := c.GlobalBool("use-geth")

	// TODO: use different db url for validator -- this helps us run both at once.
	// TODO: add validator dao to keep track of last validator point. maybe?
	// TODO: create diff directory per user.
	db, level, err := db.CreateLevelDatabase(path.Join(dburl, "validator"))

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

	go Run(c.Int("root-port"), c.Int("validator-port"), level, plasma)

	select {}
}
