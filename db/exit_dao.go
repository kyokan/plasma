package db

import (
	"github.com/syndtr/goleveldb/leveldb"
)

const latestExitIdxKey = "LATEST_EXIT_IDX"

type ExitDao interface {
	LastExitEventIdx() (uint64, error)
	SaveExitEventIdx(idx uint64) error
}

type LevelExitDao struct {
	db *leveldb.DB
}

func (dao *LevelExitDao) SaveExitEventIdx(idx uint64) error {
	gd := &GuardedDb{db: dao.db}
	key := prefixKey(latestExitIdxKey)
	b := uint64ToBytes(idx)

	gd.Put(key, b, nil)

	if gd.err != nil {
		return gd.err
	}

	return nil
}

func (dao *LevelExitDao) LastExitEventIdx() (uint64, error) {
	gd := &GuardedDb{db: dao.db}
	key := prefixKey(latestExitIdxKey)
	b := gd.Get(key, nil)

	if gd.err != nil {
		return 0, gd.err
	}

	return bytesToUint64(b), nil
}
