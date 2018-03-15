package db

import (
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"path"
)

type Database struct {
	TxDao      TransactionDao
	BlockDao   BlockDao
	MerkleDao  MerkleDao
	AddressDao AddressDao
}

func CreateLevelDatabase(location string) (*leveldb.DB, *Database, error) {
	loc := path.Join(location, "db")
	log.Printf("Creating database in %s.", loc)
	level, err := leveldb.OpenFile(loc, nil)

	if err != nil {
		return nil, nil, err
	}

	txDao := LevelTransactionDao{db: level}
	blockDao := LevelBlockDao{db: level}
	merkleDao := LevelMerkleDao{db: level}
	addressDao := LevelAddressDao{db: level, txDao: &txDao}

	return level, &Database{
		TxDao:      &txDao,
		BlockDao:   &blockDao,
		MerkleDao:  &merkleDao,
		AddressDao: &addressDao,
	}, nil
}
