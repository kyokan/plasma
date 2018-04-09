package validator

import (
	"fmt"

	"github.com/urfave/cli"
)

func Start(c *cli.Context) {
	fmt.Println("Validator Starting")

	go Run(c.Int("rpc-port"))

	select {}
}
