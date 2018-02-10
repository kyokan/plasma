package db

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
	"github.com/syndtr/goleveldb/leveldb"
	"strconv"
)

const blockKeyPrefix = "blk"

const latestKey = "LATEST_BLOCK"

type BlockDao interface {
	BlockAtHeight(num uint64) (*chain.Block, error)
	Save(blk *chain.Block) error
	Latest() (*chain.Block, error)
}

type LevelBlockDao struct {
	db     *leveldb.DB
	height uint64
}

type guardedDb struct {
	db  *leveldb.DB
	err error
}

func (dao *LevelBlockDao) Save(blk *chain.Block) error {
	cbor, err := blk.ToCbor()

	if err != nil {
		return err
	}

	key := blockPrefixKey(common.ToHex(blk.BlockHash))

	gd := &GuardedDb{db: dao.db}
	gd.Put(key, cbor, nil)
	gd.Put(blockPrefixKey(latestKey), key, nil)
	gd.Put(blockNumKey(blk.Header.Number), key, nil)

	if gd.err != nil {
		return err
	}

	dao.height = blk.Header.Number

	return nil
}

func (dao *LevelBlockDao) Latest() (*chain.Block, error) {
	key := blockPrefixKey(latestKey)

	exists, err := dao.db.Has(key, nil)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	gd := &GuardedDb{db: dao.db}
	topKey := gd.Get(key, nil)
	data := gd.Get(topKey, nil)

	if err != nil {
		return nil, err
	}

	blk, err := chain.BlockFromCbor(data)

	if err != nil {
		return nil, err
	}

	return blk, nil
}

func (dao *LevelBlockDao) BlockAtHeight(num uint64) (*chain.Block, error) {
	gd := &GuardedDb{db: dao.db}
	key := gd.Get(blockNumKey(num), nil)
	data := gd.Get(key, nil)

	if gd.err != nil {
		return nil, gd.err
	}

	blk, err := chain.BlockFromCbor(data)

	if err != nil {
		return nil, err
	}

	return blk, nil
}

func (dao *LevelBlockDao) blockHeight() (uint64, error) {
	if dao.height != 0 {
		return dao.height, nil
	}

	latest, err := dao.Latest()

	if err != nil {
		return 0, err
	}

	dao.height = latest.Header.Number
	return dao.height, nil
}

func blockNumKey(num uint64) []byte {
	return blockPrefixKey(strconv.FormatUint(num, 10))
}

func blockPrefixKey(parts ...string) []byte {
	return prefixKey(blockKeyPrefix, parts...)
}
