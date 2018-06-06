package db

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/util"
	"github.com/syndtr/goleveldb/leveldb"
	levelutil "github.com/syndtr/goleveldb/leveldb/util"
	"math/big"
	"github.com/ethereum/go-ethereum/rlp"
)

type AddressDao interface {
	Balance(addr *common.Address) (*big.Int, error)
	SpendableTxs(addr *common.Address) ([]chain.Transaction, error)
	UTXOs(addr *common.Address) ([]chain.Transaction, error)
}

type LevelAddressDao struct {
	db    *leveldb.DB
	txDao TransactionDao
}

func (dao *LevelAddressDao) Balance(addr *common.Address) (*big.Int, error) {
	txs, err := dao.SpendableTxs(addr)

	if err != nil {
		return nil, err
	}

	total := big.NewInt(0)

	for _, tx := range txs {
		total = total.Add(total, extractAmount(&tx, addr))
	}

	return total, nil
}

func (dao *LevelAddressDao) SpendableTxs(addr *common.Address) ([]chain.Transaction, error) {
	prefix := earnKey(addr)
	iter := dao.db.NewIterator(levelutil.BytesPrefix(prefix), nil)

	earnedMap := make(map[string]*chain.Flow)

	for iter.Next() {
		var flow chain.Flow
		err := rlp.DecodeBytes(iter.Value(), &flow)

		if err != nil {
			return nil, err
		}

		earnedMap[common.ToHex(flow.Hash)] = &flow
	}

	prefix = spendKey(addr)
	iter = dao.db.NewIterator(levelutil.BytesPrefix(prefix), nil)

	for iter.Next() {
		var flow chain.Flow
		err := rlp.DecodeBytes(iter.Value(), &flow)

		if err != nil {
			return nil, err
		}

		hash := common.ToHex(flow.Hash)
		if _, exists := earnedMap[hash]; exists {
			delete(earnedMap, hash)
		}
	}

	var ret []chain.Transaction
	for _, flow := range earnedMap {
		tx, err := dao.txDao.FindByBlockNumTxIdx(flow.BlkNum, flow.TxIdx)

		if err != nil {
			return nil, err
		}

		ret = append(ret, *tx)
	}

	return ret, nil
}

func (dao *LevelAddressDao) UTXOs(addr *common.Address) ([]chain.Transaction, error) {
	txs, err := dao.SpendableTxs(addr)

	if err != nil {
		return nil, err
	}

	var ret []chain.Transaction

	for _, tx := range txs {
		utxo := tx.OutputFor(addr)

		if !util.AddressesEqual(&utxo.NewOwner, addr) {
			continue
		}

		ret = append(ret, tx)
	}

	return ret, nil
}

func extractAmount(tx *chain.Transaction, addr *common.Address) *big.Int {
	return tx.OutputFor(addr).Amount
}
