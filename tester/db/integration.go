package db

import (
	"log"
	"reflect"

	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"github.com/urfave/cli"
)

func IntegrationTest(c *cli.Context) {
	dburl := c.GlobalString("db")

	db, level, err := db.CreateLevelDatabase(dburl)

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	header := &chain.BlockHeader{
		Number: 1,
	}

	lastBlock := &chain.Block{
		Header:    header,
		BlockHash: header.Hash(),
	}

	err = level.BlockDao.Save(lastBlock)

	if err != nil {
		panic(err)
	}

	res, err := level.BlockDao.BlockAtHeight(1)

	if err != nil {
		panic(err)
	}

	assert(reflect.DeepEqual(res.Header, header))
	assert(reflect.DeepEqual(res.BlockHash, header.Hash()))
}

func assert(result bool) {
	if !result {
		panic("Assert failed!")
	}
}
