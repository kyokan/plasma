package db

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/util"
	"github.com/syndtr/goleveldb/leveldb"
)

const invalidKeyPrefix = "invalid"

type InvalidBlockDao interface {
	Get(blkHash util.Hash) (*chain.Block, error)
	Save(blk *chain.Block) error
}

type LevelInvalidBlockDao struct {
	db *leveldb.DB
}

func (dao *LevelInvalidBlockDao) Save(blk *chain.Block) error {
	cbor, err := blk.ToCbor()

	if err != nil {
		return err
	}

	key := invalidPrefixKey(common.ToHex(blk.BlockHash))

	gd := &GuardedDb{db: dao.db}
	gd.Put(key, cbor, nil)

	if gd.err != nil {
		return err
	}

	return nil
}

func (dao *LevelInvalidBlockDao) Get(blkHash util.Hash) (*chain.Block, error) {
	key := invalidPrefixKey(common.ToHex(blkHash))
	gd := &GuardedDb{db: dao.db}
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

func invalidPrefixKey(parts ...string) []byte {
	return prefixKey(invalidKeyPrefix, parts...)
}
