package db

import (
	"errors"
	"log"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/util"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/ethereum/go-ethereum/rlp"
)

const txKeyPrefix = "tx"
const earnKeyPrefix = "earn"
const spendKeyPrefix = "spend"

type TransactionDao interface {
	Save(tx *chain.Transaction) error
	SaveMany(txs []chain.Transaction) error
	FindByBlockNum(blkNum uint64) ([]chain.Transaction, error)
	FindByBlockNumTxIdx(blkNum uint64, txIdx uint32) (*chain.Transaction, error)
}

type LevelTransactionDao struct {
	db *leveldb.DB
}

func (dao *LevelTransactionDao) Save(tx *chain.Transaction) error {
	return dao.SaveMany([]chain.Transaction{*tx})
}

func (dao *LevelTransactionDao) SaveMany(txs []chain.Transaction) error {
	batch := new(leveldb.Batch)

	for _, tx := range txs {
		// TODO: there is a bug here because the prev tx must always exist,
		// but it could be included in this save.
		err := dao.save(batch, &tx)

		if err != nil {
			return err
		}
	}

	return dao.db.Write(batch, nil)
}

// Currently only returns the first 100 transactions.
func (dao *LevelTransactionDao) FindByBlockNum(blkNum uint64) ([]chain.Transaction, error) {
	var txs []chain.Transaction
	gd := &GuardedDb{db: dao.db}

	for i := 0; i < 100; i++ {
		key := blkNumTxIdxKey(blkNum, uint32(i))

		data, err := gd.db.Get(key, nil)

		if err != nil {
			// Return early if we can't find the next transaction.
			if err.Error() == "leveldb: not found" {
				break
			}

			return nil, err
		}

		tx := chain.Transaction{}
		err = rlp.DecodeBytes(data, &tx)

		if err != nil {
			return nil, err
		}

		txs = append(txs, tx)
	}

	return txs, nil
}

func (dao *LevelTransactionDao) FindByBlockNumTxIdx(blkNum uint64, txIdx uint32) (*chain.Transaction, error) {
	key := blkNumTxIdxKey(blkNum, txIdx)
	exists, err := dao.db.Has(key, nil)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	gd := &GuardedDb{db: dao.db}
	data := gd.Get(key, nil)

	if gd.err != nil {
		return nil, gd.err
	}

	tx := chain.Transaction{}
	err = rlp.DecodeBytes(data, &tx)

	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func (dao *LevelTransactionDao) FindPreviousTx(tx *chain.Transaction, inputIdx uint8) (*chain.Transaction, error) {
	var input *chain.Input

	if inputIdx != 0 && inputIdx != 1 {
		panic("inputIdx must be 0 or 1")
	}

	if inputIdx == 0 {
		input = tx.Input0
	} else {
		input = tx.Input1
	}

	prevTx, err := dao.FindByBlockNumTxIdx(input.BlkNum, input.TxIdx)

	if err != nil {
		return nil, err
	}

	return prevTx, nil
}

func (dao *LevelTransactionDao) save(batch *leveldb.Batch, tx *chain.Transaction) error {
	cbor, err := rlp.EncodeToBytes(tx)

	if err != nil {
		return err
	}

	hash := tx.Hash()
	hexHash := common.ToHex(hash)
	hashKey := txPrefixKey("hash", hexHash)

	batch.Put(hashKey, cbor)
	batch.Put(blkNumHashkey(tx.BlkNum, hexHash), cbor)
	batch.Put(blkNumTxIdxKey(tx.BlkNum, tx.TxIdx), cbor)

	if tx.IsDeposit() {
		flow := chain.NewFlow(tx.BlkNum, 0, 0)
		flowEnc, err := rlp.EncodeToBytes(&flow)

		if err != nil {
			return err
		}

		batch.Put(earnKey(&tx.Output0.NewOwner), flowEnc)
		return nil
	}

	if err = dao.recordEarns(batch, tx); err != nil {
		return err
	}

	if err = dao.recordSpends(batch, tx); err != nil {
		return err
	}

	return nil
}

func (dao *LevelTransactionDao) recordEarns(batch *leveldb.Batch, tx *chain.Transaction) error {
	err := dao.recordEarn(batch, tx, 0)

	if err != nil {
		return err
	}

	if tx.Output1.IsZeroOutput() {
		return nil
	}

	err = dao.recordEarn(batch, tx, 1)

	if err != nil {
		return err
	}

	return nil
}

func (dao *LevelTransactionDao) recordEarn(batch *leveldb.Batch, tx *chain.Transaction, outIdx uint8) error {
	flow := chain.NewFlow(tx.BlkNum, tx.TxIdx, outIdx)
	flowEnc, err := rlp.EncodeToBytes(&flow)

	if err != nil {
		return err
	}

	output := tx.OutputAt(outIdx)
	batch.Put(earnKey(&output.NewOwner), flowEnc)
	return nil
}

func (dao *LevelTransactionDao) recordSpends(batch *leveldb.Batch, tx *chain.Transaction) error {
	err := dao.recordSpend(batch, tx, 0)

	if err != nil {
		return err
	}

	if tx.Input1.IsZeroInput() {
		return nil
	}

	err = dao.recordSpend(batch, tx, 1)

	if err != nil {
		return err
	}

	return nil
}

func (dao *LevelTransactionDao) recordSpend(batch *leveldb.Batch, tx *chain.Transaction, inputIdx uint8) error {
	prevTx, err := dao.FindPreviousTx(tx, inputIdx)

	if err != nil {
		return err
	}

	if prevTx == nil {
		return errors.New("expected to find an output")
	}

	input := tx.InputAt(inputIdx)
	log.Printf("%d %d %d", input.BlkNum, input.TxIdx, input.OutIdx)
	flow := chain.NewFlow(input.BlkNum, input.TxIdx, input.OutIdx)
	flowEnc, err := rlp.EncodeToBytes(&flow)

	if err != nil {
		return err
	}

	batch.Put(spendKey(&prevTx.OutputAt(input.OutIdx).NewOwner), flowEnc)

	return nil
}

func earnKey(addr *common.Address) []byte {
	return prefixKey(earnKeyPrefix, util.AddressToHex(addr))
}

func spendKey(addr *common.Address) []byte {
	return prefixKey(spendKeyPrefix, util.AddressToHex(addr))
}

func blkNumHashkey(blkNum uint64, hexHash string) []byte {
	return txPrefixKey("blkNum", strconv.FormatUint(blkNum, 10), "hash", hexHash)
}

func blkNumTxIdxKey(blkNum uint64, txIdx uint32) []byte {
	return txPrefixKey("blkNum", strconv.FormatUint(blkNum, 10), "txIdx", strconv.FormatUint(uint64(txIdx), 10))
}

func txPrefixKey(parts ...string) []byte {
	return prefixKey(txKeyPrefix, parts...)
}
