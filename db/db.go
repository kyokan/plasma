package db

import (
	"log"
	"path"

	"github.com/syndtr/goleveldb/leveldb"
)

type Database struct {
	TxDao           TransactionDao
	BlockDao        BlockDao
	MerkleDao       MerkleDao
	AddressDao      AddressDao
	DepositDao      DepositDao
	ExitDao         ExitDao
	InvalidBlockDao InvalidBlockDao
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
	depositDao := LevelDepositDao{db: level}
	exitDao := LevelExitDao{db: level}
	invalidBlockDao := LevelInvalidBlockDao{db: level}

	return level, &Database{
		TxDao:           &txDao,
		BlockDao:        &blockDao,
		MerkleDao:       &merkleDao,
		AddressDao:      &addressDao,
		DepositDao:      &depositDao,
		ExitDao:         &exitDao,
		InvalidBlockDao: &invalidBlockDao,
	}, nil
}
