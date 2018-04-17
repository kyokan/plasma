package db

import (
	"fmt"
	"strconv"

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
		// TODO: if it's not found return 0.
		return 0, gd.err
	}

	return bytesToUint64(b), nil
}

func uint64ToBytes(i uint64) []byte {
	return []byte(fmt.Sprintf("%X", i))
}

func bytesToUint64(b []byte) uint64 {
	s := string(b)

	n, err := strconv.ParseUint(s, 16, 32)

	if err != nil {
		panic(err)
	}

	return uint64(n)
}
