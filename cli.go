package main

import (
	"fmt"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/plasma"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "db",
			Value: db.DefaultLocation(),
			Usage: "Filepath for Plasma's database.",
		},
		cli.StringFlag{
			Name:  "node-url",
			Value: "http://localhost:8545",
			Usage: "Full URL to a running geth node.",
		},
		cli.StringFlag{
			Name:  "contract-addr",
			Value: "0xd1d7dddd82189ea452eb5e104d13f0ca367887d9",
			Usage: "Plasma contract address.",
		},
	}

	app.Name = "Plasma"
	app.Usage = "Demonstrates what an example Plasma blockchain can do."
	app.Commands = []cli.Command{
		{
			Name:   "start",
			Usage:  "Starts running a Plasma root node.",
			Action: plasma.Start,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "rpc-port",
					Value: 8643,
					Usage: "Port for the RPC server to listen on.",
				},
			},
		},
		{
			Name:   "validate",
			Usage:  "Starts running a Plasma validator node.",
			Action: func() { fmt.Println("Not implemented yet.") },
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "root-url",
					Usage: "The URL of the root node.",
				},
			},
		},
		{
			Name:   "utxos",
			Usage:  "Prints UTXOs for the given address.",
			Action: plasma.PrintUTXOs,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "addr",
					Usage: "The address to print UTXOs for.",
				},
			},
		},
	}

	app.Run(os.Args)
}
