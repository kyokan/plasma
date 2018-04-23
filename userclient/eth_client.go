package userclient

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/eth"
	"github.com/urfave/cli"
)

func GetBalance(c *cli.Context) {
	nodeURL := c.GlobalString("node-url")
	userAddress := c.GlobalString("user-address")
	client, err := eth.NewClient(nodeURL)

	if err != nil {
		log.Fatalf("Failed to create eth client: %v", err)
	}

	addr := common.HexToAddress(userAddress)

	res, err := client.GetBalance(addr)

	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}

	fmt.Printf("Balance: %v\n", res)
}
