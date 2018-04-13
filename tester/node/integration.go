package node

import (
	"log"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/node"
	"github.com/kyokan/plasma/util"
	"github.com/urfave/cli"
)

func IntegrationTest(c *cli.Context) {
	contractAddress := c.GlobalString("contract-addr")
	nodeURL := c.GlobalString("node-url")
	keystoreDir := c.GlobalString("keystore-dir")
	keystoreFile := c.GlobalString("keystore-file")
	userAddress := c.GlobalString("user-address")
	privateKey := c.GlobalString("private-key")
	signPassphrase := c.GlobalString("sign-passphrase")
	dburl := c.GlobalString("db")

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
	plasma := eth.CreatePlasmaClient(nodeURL, contractAddress, userAddress, privateKeyECDSA)

	p := node.NewPlasmaNode(level, sink, plasma)

	go p.Start()

	// go func() {
	// 	chch := make(chan chan node.TransactionRequest)
	// 	go rpc.Start(c.Int("rpc-port"), chch)
	// 	sink.AcceptTransactionRequests(chch)
	// }()

}
