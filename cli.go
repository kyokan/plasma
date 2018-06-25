package main

import (
	"os"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/plasma"
	db_tests "github.com/kyokan/plasma/tester/db"
	plasma_tests "github.com/kyokan/plasma/tester/plasma"
	pq_tests "github.com/kyokan/plasma/tester/pq"
	"github.com/kyokan/plasma/userclient"
	"github.com/kyokan/plasma/validator"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/urfave/cli.v1/altsrc"
)

func main() {
	app := cli.NewApp()

	flags := []cli.Flag{
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "config",
			Usage: "Filepath for Plasma's configuration.",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "db",
			Value: db.DefaultLocation(),
			Usage: "Filepath for Plasma's database.",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "node-url",
			Usage: "Full URL to a running geth node.",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "contract-addr",
			Usage: "Plasma contract address.",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "priority-queue-contract-addr",
			Usage: "Plasma contract address.",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "keystore-dir",
			Usage: "Keystore directory.",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "keystore-file",
			Usage: "Keystore file.",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "user-address",
			Usage: "User address.",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "private-key",
			Usage: "Private key of user address.",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "sign-passphrase",
			Usage: "Passphrase for keystore file.",
		}),
		altsrc.NewBoolFlag(cli.BoolFlag{
			Name:  "use-geth",
			Usage: "Use geth to sign transactions.",
		}),
	}

	loadCfgFn := func(context *cli.Context) (altsrc.InputSourceContext, error) {
		cfgFilePath := context.String("config")
		if cfgFilePath != "" {
			return altsrc.NewYamlSourceFromFile(cfgFilePath)
		}
		return &altsrc.MapInputSource{}, nil
	}

	app.Before = altsrc.InitInputSourceWithContext(flags, loadCfgFn)
	app.Flags = flags

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
					Name:  "root-port",
					Value: 8643,
					Usage: "Port for the root server to listen on.",
				},
				cli.IntFlag{
					Name:  "validator-port",
					Value: 8644,
					Usage: "Port for the validator server to listen on.",
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
			Name:   "deposit",
			Usage:  "Runs deposit.",
			Action: userclient.Deposit,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "amount",
					Usage: "Amount to deposit.",
				},
			},
		},
		{
			Name:   "exit",
			Usage:  "Runs exit started",
			Action: userclient.StartExit,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "root-port",
					Value: 8643,
					Usage: "Port for the root server to listen on.",
				},
				cli.IntFlag{
					Name:  "blocknum",
					Usage: "Block to exit.",
				},
				cli.IntFlag{
					Name:  "txindex",
					Usage: "Transaction to exit.",
				},
				cli.IntFlag{
					Name:  "oindex",
					Usage: "Output to exit.",
				},
			},
		},
		{
			Name:   "finalize",
			Usage:  "Runs finalize",
			Action: userclient.Finalize,
		},
		{
			Name:   "balance",
			Usage:  "Runs get balance",
			Action: userclient.GetBalance,
		},
		{
			Name:   "block",
			Usage:  "Runs get blocks",
			Action: userclient.GetBlockCLI,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "root-port",
					Value: 8643,
					Usage: "Port for the root server to listen on.",
				},
				cli.IntFlag{
					Name:  "height",
					Usage: "Block height.",
				},
			},
		},
		{
			Name:   "send",
			Usage:  "Runs send",
			Action: userclient.SendCLI,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "root-port",
					Value: 8643,
					Usage: "Port for the root server to listen on.",
				},
				cli.StringFlag{
					Name:  "to",
					Usage: "Recipient.",
				},
				cli.IntFlag{
					Name:  "amount",
					Usage: "Amont to send.",
				},
			},
		},
		{
			Name:   "force-submit",
			Usage:  "Runs force submit block",
			Action: plasma.ForceSubmitBlock,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "merkle-root",
					Usage: "Merkle root.",
				},
				cli.StringFlag{
					Name:  "prev-hash",
					Usage: "Previous block hash.",
				},
				cli.IntFlag{
					Name:  "number",
					Usage: "Block number.",
				},
			},
		},
	}

	app.Run(os.Args)
}
