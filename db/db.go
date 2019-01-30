package db

import (
	"log"
	"path"

		"github.com/syndtr/goleveldb/leveldb"
)

func CreateStorage(location string) (*leveldb.DB, PlasmaStorage, error) {
	loc := path.Join(location, "db")
	log.Printf("Creating database in %s.", loc)
	level, err := leveldb.OpenFile(loc, nil)
	if err != nil {
		return nil, nil, err
	}
	return level, NewStorage(level), nil
}