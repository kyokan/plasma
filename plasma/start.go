package plasma

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/node"
	"github.com/kyokan/plasma/rpc"
	"github.com/urfave/cli"
	"log"
)

func Start(c *cli.Context) {
	level, err := db.CreateLevelDatabase(c.GlobalString("db"))

	if err != nil {
		log.Panic(err)
	}

	client, err := eth.NewClient(c.GlobalString("node-url"))

	if err != nil {
		log.Panic("Failed to start ETH client: ", err)
	}

	sink := node.NewTransactionSink(level, client)
	p := node.NewPlasmaNode(level, sink)

	go p.Start()

	go func() {
		chch := make(chan chan node.TransactionRequest)
		go rpc.Start(c.Int("rpc-port"), chch)
		sink.AcceptTransactionRequests(chch)
	}()

	go func() {

		ch := make(chan eth.DepositEvent)
		err = client.SubscribeDeposits(common.HexToAddress(c.GlobalString("contract-addr")), ch)

		if err != nil {
			log.Panic("Failed to subscribe to deposits: ", err)
		}

		sink.AcceptDepositEvents(ch)
	}()

	select {}
}
