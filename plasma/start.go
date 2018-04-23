package plasma

import (
	"log"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/node"
	"github.com/kyokan/plasma/rpc"
	"github.com/urfave/cli"
)

func Start(c *cli.Context) {
	nodeURL := c.GlobalString("node-url")
	dburl := c.GlobalString("db")

	plasma := eth.CreatePlasmaClientCLI(c)

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

	p := node.NewPlasmaNode(level, sink, plasma)

	go p.Start()

	go rpc.Start(c.Int("rpc-port"), level, sink)

	// TODO: ensure that 1 deposit tx is always 1 block
	go node.StartDepositListener(level, sink, plasma)

	// TODO: add an exit listener to make sure to add an exit transaction to root node.
	// Also add an exit block to the plasma contract.

	select {}
}
