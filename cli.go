package main

import (
	"os"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/plasma"
	db_tests "github.com/kyokan/plasma/tester/db"
	plasma_tests "github.com/kyokan/plasma/tester/plasma"
	pq_tests "github.com/kyokan/plasma/tester/pq"
	"github.com/kyokan/plasma/validator"
	"github.com/urfave/cli"
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
			Name: "node-url",
			// Value: "http://localhost:8545", // dev
			// Value: "ws://localhost:8546", // websocket
			Value: "http://localhost:7545",
			Usage: "Full URL to a running geth node.",
		},
		cli.StringFlag{
			Name: "contract-addr",
			// Value: "0xd1d7dddd82189ea452eb5e104d13f0ca367887d9", // test
			// Value: "0x4db27d728a8714af06474786dbaeadea9673c511", / dev
			Value: "0xf25186b5081ff5ce73482ad761db0eb0d25abfbf",
			Usage: "Plasma contract address.",
		},
		cli.StringFlag{
			Name:  "priority-queue-contract-addr",
			Value: "0xecfcab0a285d3380e488a39b4bb21e777f8a4eac",
			Usage: "Plasma contract address.",
		},
		cli.StringFlag{
			Name:  "keystore-dir",
			Value: "/Users/mattkim/geth/chain/keystore", // private chain
			Usage: "Keystore directory.",
		},
		cli.StringFlag{
			Name:  "keystore-file",
			Value: "/Users/mattkim/geth/chain/keystore/UTC--2018-03-13T17-33-34.839516799Z--44a5cae1ebd47c415630da1e2131b71d1f2f5803", // private chain
			Usage: "Keystore file.",
		},
		cli.StringFlag{
			Name: "user-address",
			// Value: "44a5cae1ebd47c415630da1e2131b71d1f2f5803" // private chain
			Value: "0x627306090abaB3A6e1400e9345bC60c78a8BEf57", // ganache
			Usage: "User address.",
		},
		cli.StringFlag{
			Name:  "private-key",
			Value: "c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3", // ganache
			Usage: "Private key of user address.",
		},
		cli.StringFlag{
			Name:  "sign-passphrase",
			Value: "test", // private chain
			Usage: "Passphrase for keystore file.",
		},
	}

	app.Name = "Plasma"
	app.Usage = "A secure and scalable solution for decentralized applications."
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
			Action: validator.Start,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "rpc-port",
					Value: 8644,
					Usage: "Port for the RPC server to listen on.",
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
		{
			Name:   "plasma-tests",
			Usage:  "Runs plasma integration tests.",
			Action: plasma_tests.IntegrationTest,
		},
		{
			Name:   "pq-tests",
			Usage:  "Runs priority queue integration tests.",
			Action: pq_tests.IntegrationTest,
		},
		{
			Name:   "db-tests",
			Usage:  "Runs level db integration tests.",
			Action: db_tests.IntegrationTest,
		},
		{
			Name:   "validator-main",
			Usage:  "Runs validator main.",
			Action: validator.Main,
		},
	}

	app.Run(os.Args)
}
