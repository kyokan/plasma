package db

import (
	"github.com/syndtr/goleveldb/leveldb"
)

const latestDepositIdxKey = "LATEST_DEPOSIT_IDX"

type DepositDao interface {
	LastDepositEventIdx() (uint64, error)
	SaveDepositEventIdx(idx uint64) error
}

type LevelDepositDao struct {
	db *leveldb.DB
}

func (dao *LevelDepositDao) SaveDepositEventIdx(idx uint64) error {
	gd := &GuardedDb{db: dao.db}
	key := prefixKey(latestDepositIdxKey)
	b := uint64ToBytes(idx)

	gd.Put(key, b, nil)

	if gd.err != nil {
		return gd.err
	}

	return nil
}

func (dao *LevelDepositDao) LastDepositEventIdx() (uint64, error) {
	gd := &GuardedDb{db: dao.db}
	key := prefixKey(latestDepositIdxKey)
	b := gd.Get(key, nil)

	if gd.err != nil {
		return 0, gd.err
	}

	return bytesToUint64(b), nil
}
