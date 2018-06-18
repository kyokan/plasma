package db

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/util"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/ethereum/go-ethereum/rlp"
)

const merkleKeyPrefix = "merkle"

type MerkleDao interface {
	Save(n *util.MerkleNode) error
	SaveMany(ns []util.MerkleNode) error
}

type LevelMerkleDao struct {
	db *leveldb.DB
}

func (dao *LevelMerkleDao) Save(n *util.MerkleNode) error {
	return dao.SaveMany([]util.MerkleNode{*n})
}

func (dao *LevelMerkleDao) SaveMany(ns []util.MerkleNode) error {
	batch := new(leveldb.Batch)

	for _, n := range ns {
		enc, err := rlp.EncodeToBytes(&n)

		if err != nil {
			return err
		}

		batch.Put(merklePrefixKey(common.ToHex(n.Hash)), enc)
	}

	return dao.db.Write(batch, nil)
}

func merklePrefixKey(parts ...string) []byte {
	return prefixKey(merkleKeyPrefix, parts...)
}
