package plasma

import (
	"log"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/node"
	"github.com/kyokan/plasma/rpc"
	"github.com/kyokan/plasma/util"
	"github.com/urfave/cli"
)

func Start(c *cli.Context) {
	contractAddress := c.GlobalString("contract-addr")
	nodeURL := c.GlobalString("node-url")
	keystoreDir := c.GlobalString("keystore-dir")
	keystoreFile := c.GlobalString("keystore-file")
	userAddress := c.GlobalString("user-address")
	privateKey := c.GlobalString("private-key")
	signPassphrase := c.GlobalString("sign-passphrase")
	dburl := c.GlobalString("db")
	useGeth := c.GlobalBool("use-geth")

	db, level, err := db.CreateLevelDatabase(dburl)

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	client, err := eth.NewClient(nodeURL)

	if err != nil {
		log.Panic("Failed to start ETH client: ", err)
	}

	sink := node.NewTransactionSink(level, client)

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

	p := node.NewPlasmaNode(level, sink, plasma)

	go p.Start()

	go rpc.Start(c.Int("rpc-port"), level, sink)

	// TODO: deposits must always be in their own block.
	go node.StartDepositListener(level, sink, plasma)

	// TODO: add an exit listener to make sure to add an exit transaction.
	// inputs linking to the utxo, and outputs with zeros.
	// This also needs to match the tx that is added to the plasma chain.

	select {}
}

func exists(s string) bool {
	return len(s) != 0
}
