package validator

import (
	"fmt"
	"log"
	"path"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"gopkg.in/urfave/cli.v1"
)

func Start(c *cli.Context) {
	log.Println("Validator Starting")

	userAddress := c.GlobalString("user-address")
	dburl := c.GlobalString("db")
	// TODO: turn this into a client.
	rootUrl := fmt.Sprintf("http://localhost:%d/rpc", c.Int("root-port"))
	validatorPort := c.Int("validator-port")

	plasma := eth.CreatePlasmaClientCLI(c)

	db, storage, err := db.CreateStorage(path.Join(dburl, "validator", userAddress), plasma)

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	go RootNodeListener(rootUrl, storage, plasma, userAddress)

	go ExitStartedListener(rootUrl, storage, plasma)

	go Run(validatorPort)

	select {}
}
