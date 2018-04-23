package validator

import (
	"log"
	"path"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/urfave/cli"
)

func Start(c *cli.Context) {
	log.Println("Validator Starting")

	dburl := c.GlobalString("db")

	plasma := eth.CreatePlasmaClientCLI(c)

	// TODO: create diff directory per user.
	db, level, err := db.CreateLevelDatabase(path.Join(dburl, "validator"))

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	go Run(c.Int("root-port"), c.Int("validator-port"), level, plasma)

	select {}
}
