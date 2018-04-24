package validator

import (
	"fmt"
	"log"
	"path"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/urfave/cli"
)

func Start(c *cli.Context) {
	log.Println("Validator Starting")

	userAddress := c.GlobalString("user-address")
	dburl := c.GlobalString("db")
	// TODO: turn this into a client.
	rootUrl := fmt.Sprintf("http://localhost:%d/rpc", c.Int("root-port"))
	validatorPort := c.Int("validator-port")

	plasma := eth.CreatePlasmaClientCLI(c)

	db, level, err := db.CreateLevelDatabase(path.Join(dburl, "validator", userAddress))

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	go RootNodeListener(rootUrl, level, plasma, userAddress)

	go ExitStartedListener(rootUrl, level, plasma)

	go Run(validatorPort)

	select {}
}
