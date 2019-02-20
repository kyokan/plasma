package db

import (
	"github.com/kyokan/plasma/pkg/chain"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kyokan/plasma/util"
	"github.com/kyokan/plasma/pkg/merkle"
	"github.com/golang/protobuf/proto"
	"time"
	"strings"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"errors"
	levelutil "github.com/syndtr/goleveldb/leveldb/util"
)

type LevelStorage struct {
	db *leveldb.DB
}

func NewLevelStorage(db *leveldb.DB) Storage {
	result := &LevelStorage{
		db: db,
	}

	return result
}

func (ps *LevelStorage) findPreviousTx(tx *chain.Transaction, inputIdx uint8) (*chain.ConfirmedTransaction, error) {
	var input *chain.Input
	if inputIdx != 0 && inputIdx != 1 {
		panic("inputIdx must be 0 or 1")
	}

	if inputIdx == 0 {
		input = tx.Body.Input0
	} else {
		input = tx.Body.Input1
	}

	return ps.findTransactionByBlockNumTxIdx(input.BlockNumber, input.TransactionIndex)
}

func (ps *LevelStorage) saveTransaction(blockNum uint64, txIdx uint32, originalTx chain.Transaction, batch *leveldb.Batch) (*chain.ConfirmedTransaction, error) {
	tx := originalTx.Clone()
	tx.Body.TransactionIndex = txIdx
	tx.Body.BlockNumber = blockNum

	confirmed := &chain.ConfirmedTransaction{
		Transaction: tx,
	}
	if err := ps.batchWriteConfirmedTransaction(confirmed, batch); err != nil {
		return nil, err
	}

	return confirmed, nil
}

func (ps *LevelStorage) batchWriteConfirmedTransaction(confirmed *chain.ConfirmedTransaction, batch *leveldb.Batch) error {
	tx := confirmed.Transaction
	confirmedEnc, err := proto.Marshal(confirmed.Proto())
	if err != nil {
		return err
	}
	hash := confirmed.Hash()
	hexHash := hexutil.Encode(hash)

	batch.Put(txByHashKey(hexHash), confirmedEnc)
	batch.Put(txByBlockNumTxIdxKey(tx.Body.BlockNumber, tx.Body.TransactionIndex), []byte(hexHash))

	var empty []byte

	// Recording spends
	if tx.Body.IsDeposit() {
		batch.Put(depositKey(tx.Body.Input0.DepositNonce), []byte(hexHash))
	} else {
		prevConfirmed0, err := ps.findPreviousTx(tx, 0)
		if err != nil {
			return err
		}

		prevTx := prevConfirmed0.Transaction
		batch.Put(spendByTxIdxKey(prevTx.Body.BlockNumber, prevTx.Body.TransactionIndex, tx.Body.Input0.OutputIndex), []byte(hexHash))
		batch.Delete(utxoKey(prevTx.Body.OutputAt(tx.Body.Input0.OutputIndex).Owner, prevConfirmed0.Hash(), tx.Body.Input0.OutputIndex))
		// TODO mark as exited
	}

	if !tx.Body.Input1.IsZero() {
		prevConfirmed1, err := ps.findPreviousTx(tx, 1)
		if err != nil {
			return err
		}

		prevTx := prevConfirmed1.Transaction
		batch.Put(spendByTxIdxKey(prevTx.Body.BlockNumber, prevTx.Body.TransactionIndex, tx.Body.Input1.OutputIndex), []byte(hexHash))
		batch.Delete(utxoKey(prevTx.Body.OutputAt(tx.Body.Input0.OutputIndex).Owner, prevConfirmed1.Hash(), tx.Body.Input1.OutputIndex))
	}

	// Recording earns
	if !tx.Body.Output0.IsZeroOutput() {
		output := tx.Body.OutputAt(0)
		batch.Put(utxoKey(output.Owner, hash, 0), empty)
	}
	if !tx.Body.Output1.IsZeroOutput() {
		output := tx.Body.OutputAt(1)
		batch.Put(utxoKey(output.Owner, hash, 1), empty)
	}

	return nil
}

func (ps *LevelStorage) MarkTransactionAsExited(plasmaBlockNum uint64, plasmaTxIdx uint32, outIdx uint8, ethBlockNumber uint64, ethTransactionHash common.Hash) error {
	tx, err := ps.findTransactionByBlockNumTxIdx(plasmaBlockNum, plasmaTxIdx)
	if err != nil {
		return err
	}

	loc := &ExitLocator{
		PlasmaBlockNumber: plasmaBlockNum,
		PlasmaTransactionIndex: plasmaTxIdx,
		PlasmaOutputIndex:outIdx,
		EthereumBlockNumber:ethBlockNumber,
		EthereumTransactionHash:ethTransactionHash[:],
	}
	locBytes, err := loc.MarshalBinary()
	if err != nil {
		return err
	}
	batch := new(leveldb.Batch)
	batch.Delete(utxoKey(tx.Transaction.Body.OutputAt(outIdx).Owner, tx.Hash(), outIdx))
	batch.Put(exitKey(tx.Transaction.Body.BlockNumber, tx.Transaction.Body.TransactionIndex, outIdx), locBytes)
	return ps.db.Write(batch, nil)
}

func (ps *LevelStorage) IsDoubleSpent(tx *chain.Transaction) (bool, error) {
	body := tx.Body

	if tx.Body.IsDeposit() {
		return ps.db.Has(depositKey(tx.Body.Input0.DepositNonce), nil)
	}

	prevTx0, err := ps.findPreviousTx(tx, 0)
	if err != nil {
		return false, err
	}
	searchKeys := make([][]byte, 0)

	prevTx0Body := prevTx0.Transaction.Body
	searchKeys = append(searchKeys, spendByTxIdxKey(prevTx0Body.BlockNumber, prevTx0Body.TransactionIndex, body.Input0.OutputIndex))
	searchKeys = append(searchKeys, exitKey(prevTx0Body.BlockNumber, prevTx0Body.TransactionIndex, body.Input0.OutputIndex))
	if !body.Input1.IsZero() {
		prevTx1, err := ps.findPreviousTx(tx, 1)
		if err != nil {
			return false, err
		}
		prevTx1Body := prevTx1.Transaction.Body
		searchKeys = append(searchKeys, spendByTxIdxKey(prevTx1Body.BlockNumber, prevTx1Body.TransactionIndex, body.Input1.OutputIndex))
		searchKeys = append(searchKeys, exitKey(prevTx1Body.BlockNumber, prevTx1Body.TransactionIndex, body.Input1.OutputIndex))
	}

	for _, spendKey := range searchKeys {
		found, _ := ps.db.Has(spendKey, nil)
		if found {
			return true, nil
		}
	}

	return false, nil
}

func (ps *LevelStorage) FindDoubleSpendingTransaction(blkNum uint64, txIdx uint32, outIndex uint8) (*chain.ConfirmedTransaction, error) {
	spendingHash, err := ps.db.Get(spendByTxIdxKey(blkNum, txIdx, outIndex), nil)
	if err == leveldb.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	hash, err := hexutil.Decode(string(spendingHash))
	if err != nil {
		return nil, err
	}
	return ps.findTransactionByHash(hash)
}

func (ps *LevelStorage) FindDoubleSpendingDeposit(nonce *big.Int) (*chain.ConfirmedTransaction, error) {
	spendingHash, err := ps.db.Get(depositKey(nonce), nil)
	if err == leveldb.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	hash, err := hexutil.Decode(string(spendingHash))
	if err != nil {
		return nil, err
	}
	return ps.findTransactionByHash(hash)
}

func (ps *LevelStorage) doPackageBlock(txs []chain.Transaction) (*chain.BlockResult, error) {
	prevBlock, err := ps.LatestBlock()
	if err != nil {
		return nil, err
	}

	var blkNum uint64
	var prevHash util.Hash
	if prevBlock == nil {
		blkNum = 1
	} else {
		blkNum = prevBlock.Header.Number + 1
		prevHash = prevBlock.BlockHash
	}

	// The batch will act as in-memory buffer
	batch := new(leveldb.Batch)
	numberOfTransactions := len(txs)

	hashables := make([]util.RLPHashable, numberOfTransactions)
	for i, tx := range txs {
		hashables[i] = &tx
	}
	merkleRoot := merkle.Root(hashables)

	header := chain.BlockHeader{
		MerkleRoot: merkleRoot,
		PrevHash:   prevHash,
		Number:     blkNum,
	}
	block := &chain.Block{
		Header:    &header,
		BlockHash: header.Hash(),
	}
	currentFees := big.NewInt(0)
	for i, tx := range txs {
		_, err := ps.saveTransaction(blkNum, uint32(i), tx, batch)
		if err != nil {
			return nil, err
		}

		currentFees = currentFees.Add(currentFees, tx.Body.Fee)
	}
	meta := &chain.BlockMetadata{
		CreatedAt:        uint64(time.Now().Unix()),
		TransactionCount: uint32(numberOfTransactions),
		Fees:             currentFees,
	}

	if err := ps.batchWriteBlock(block, batch); err != nil {
		return nil, err
	}
	if err := ps.batchWriteBlockMeta(block, meta, batch); err != nil {
		return nil, err
	}

	err = ps.db.Write(batch, nil)
	if err != nil {
		return nil, err
	}

	return &chain.BlockResult{
		MerkleRoot:         merkleRoot,
		NumberTransactions: uint32(numberOfTransactions),
		BlockFees:          currentFees,
		BlockNumber:        big.NewInt(int64(block.Header.Number)),
	}, nil
}

func (ps *LevelStorage) batchWriteBlock(block *chain.Block, batch *leveldb.Batch) error {
	merkleRoot := block.Header.MerkleRoot
	enc, err := rlp.EncodeToBytes(merkleRoot)
	if err != nil {
		return err
	}
	batch.Put(merklePrefixKey(hexutil.Encode(merkleRoot)), enc)

	enc, err = rlp.EncodeToBytes(block)
	if err != nil {
		return err
	}
	key := blockPrefixKey(hexutil.Encode(block.BlockHash))
	batch.Put(key, enc)
	batch.Put(blockPrefixKey(latestKey), key)
	batch.Put(blockNumKey(block.Header.Number), key)
	return nil
}

func (ps *LevelStorage) batchWriteBlockMeta(block *chain.Block, meta *chain.BlockMetadata, batch *leveldb.Batch) error {
	metaEnc, err := meta.RLP()
	if err != nil {
		return err
	}
	batch.Put(blockMetaPrefixKey(block.Header.Number), metaEnc)
	return nil
}

func (ps *LevelStorage) InsertBlock(block *chain.Block, meta *chain.BlockMetadata, txs []chain.ConfirmedTransaction) error {
	latest, err := ps.LatestBlock()
	if err != nil {
		return err
	}
	height := uint64(0)
	if latest != nil {
		height = latest.Header.Number
	}

	if block.Header.Number != height+1 {
		return errors.New("cannot insert a block more than 1 block number ahead")
	}

	batch := new(leveldb.Batch)
	if err := ps.batchWriteBlock(block, batch); err != nil {
		return err
	}
	if err := ps.batchWriteBlockMeta(block, meta, batch); err != nil {
		return err
	}
	for _, tx := range txs {
		if err := ps.batchWriteConfirmedTransaction(&tx, batch); err != nil {
			return err
		}
	}

	return ps.db.Write(batch, nil)
}

func (ps *LevelStorage) ProcessDeposit(confirmed chain.Transaction) (deposit *chain.BlockResult, err error) {
	if !confirmed.Body.IsDeposit() {
		return nil, errors.New("only deposit blocks are accepted")
	}

	return ps.doPackageBlock([]chain.Transaction{confirmed})
}

func (ps *LevelStorage) FindTransactionsByBlockNum(blockNum uint64) ([]chain.ConfirmedTransaction, error) {
	txIter := ps.db.NewIterator(levelutil.BytesPrefix(txByBlockNumIterKey(blockNum)), nil)
	defer txIter.Release()

	var ret []chain.ConfirmedTransaction

	for txIter.Next() {
		txHash, err := hexutil.Decode(string(txIter.Value()))
		if err != nil {
			return nil, err
		}
		tx, err := ps.findTransactionByHash(txHash)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *tx)
	}

	return ret, nil
}

func (ps *LevelStorage) findTransactionByDepositNonce(nonce *big.Int) (*chain.Transaction, util.Hash, error) {
	keyPrefix := depositKey(nonce)
	iter := ps.db.NewIterator(levelutil.BytesPrefix(keyPrefix), nil)
	defer iter.Release()

	for iter.Next() {
		var confirmed chain.Transaction
		key := string(iter.Key())
		keyParts := strings.Split(key, "::")
		if len(keyParts) != 4 {
			return nil, nil, errors.New(fmt.Sprintf("Failed to parse deposit transaction position from key %s", key))
		}
		blkNum, success := util.Str2Uint64(keyParts[2])
		if success == false {
			return nil, nil, errors.New(fmt.Sprintf("Failed to parse block number from key %s", key))
		}
		txIdx, success := util.Str2Uint32(keyParts[3])
		if success == false {
			return nil, nil, errors.New(fmt.Sprintf("Failed to parse transaction index from key %s", key))
		}
		err := rlp.DecodeBytes(iter.Value(), &confirmed)
		if err != nil {
			return nil, nil, err
		}
		confirmed.Body.BlockNumber = blkNum
		confirmed.Body.TransactionIndex = txIdx
		return &confirmed, nil, nil
	}
	return nil, nil, errors.New(fmt.Sprintf("Failed to find deposit for deposit nonce %s", nonce.String()))
}

func (ps *LevelStorage) findTransactionByBlockNumTxIdx(blockNum uint64, txIdx uint32) (*chain.ConfirmedTransaction, error) {
	hash, err := ps.db.Get(txByBlockNumTxIdxKey(blockNum, txIdx), nil)
	if err != nil {
		return nil, err
	}
	hashBytes, err := hexutil.Decode(string(hash))
	if err != nil {
		return nil, err
	}
	return ps.findTransactionByHash(hashBytes)
}

func (ps *LevelStorage) findTransactionByHash(hash util.Hash) (*chain.ConfirmedTransaction, error) {
	txBytes, err := ps.db.Get(txByHashKey(hexutil.Encode(hash)), nil)
	if err != nil {
		return nil, err
	}
	return chain.UnmarshalConfirmedTransactionProto(txBytes)
}

func (ps *LevelStorage) FindTransactionByBlockNumTxIdx(blkNum uint64, txIdx uint32) (*chain.ConfirmedTransaction, error) {
	tx, err := ps.findTransactionByBlockNumTxIdx(blkNum, txIdx)
	return tx, err
}

// Address
func (ps *LevelStorage) Balance(addr common.Address) (*big.Int, error) {
	txs, err := ps.SpendableTxs(addr)
	if err != nil {
		return nil, err
	}

	total := big.NewInt(0)
	for _, confirmed := range txs {
		total = total.Add(total, extractAmount(&confirmed, addr))
	}

	return total, nil
}

func (ps *LevelStorage) SpendableTxs(addr common.Address) ([]chain.ConfirmedTransaction, error) {
	utxoIter := ps.db.NewIterator(levelutil.BytesPrefix(utxoAddrIterKey(addr)), nil)
	defer utxoIter.Release()

	ret := make([]chain.ConfirmedTransaction, 0)
	for utxoIter.Next() {
		txHash, _ := utxoKeyParts(string(utxoIter.Key()))
		spendableTx, err := ps.findTransactionByHash(txHash)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *spendableTx)
	}
	sortTransactions(ret)
	return ret, nil
}

func (ps *LevelStorage) UTXOs(addr common.Address) ([]chain.ConfirmedTransaction, error) {
	utxoIter := ps.db.NewIterator(levelutil.BytesPrefix(utxoAddrIterKey(addr)), nil)
	defer utxoIter.Release()

	var ret []chain.ConfirmedTransaction
	for utxoIter.Next() {
		txHash, _ := utxoKeyParts(string(utxoIter.Key()))
		tx, err := ps.findTransactionByHash(txHash)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *tx)
	}
	sortTransactions(ret)
	return ret, nil
}

func (ps *LevelStorage) ConfirmTransaction(blockNumber uint64, transactionIndex uint32, sigs [2]chain.Signature) (*chain.ConfirmedTransaction, error) {
	tx, err := ps.findTransactionByBlockNumTxIdx(blockNumber, transactionIndex)
	if err != nil {
		return nil, err
	}

	tx.ConfirmSigs = sigs
	hexHash := hexutil.Encode(tx.Hash())
	txBytes, err := tx.MarshalProto()
	if err != nil {
		return nil, err
	}
	ps.db.Put(txByHashKey(hexHash), txBytes, nil)
	return tx, nil
}

func (ps *LevelStorage) ConfirmSigsFor(blockNumber uint64, transactionIndex uint32) ([2]chain.Signature, error) {
	var sigs [2]chain.Signature
	tx, err := ps.findTransactionByBlockNumTxIdx(blockNumber, transactionIndex)
	if err != nil {
		return sigs, err
	}

	return tx.ConfirmSigs, err
}

// Block
func (ps *LevelStorage) BlockAtHeight(num uint64) (*chain.Block, error) {
	key, err := ps.db.Get(blockNumKey(num), nil)
	if err != nil {
		return nil, err
	}
	data, err := ps.db.Get(key, nil)
	if err != nil {
		return nil, err
	}

	var blk chain.Block
	err = rlp.DecodeBytes(data, &blk)

	if err != nil {
		return nil, err
	}

	return &blk, nil
}

func (ps *LevelStorage) BlockMetaAtHeight(num uint64) (*chain.BlockMetadata, error) {
	data, err := ps.db.Get(blockMetaPrefixKey(num), nil)
	if err != nil {
		return nil, err
	}

	var meta chain.BlockMetadata
	err = meta.FromRLP(data)
	if err != nil {
		return nil, err
	}

	return &meta, nil
}

func (ps *LevelStorage) FullBlockAtHeight(num uint64) (*chain.Block, *chain.BlockMetadata, []chain.ConfirmedTransaction, error) {
	block, err := ps.BlockAtHeight(num)
	if err != nil {
		return nil, nil, nil, err
	}
	meta, err := ps.BlockMetaAtHeight(num)
	if err != nil {
		return nil, nil, nil, err
	}
	txs, err := ps.FindTransactionsByBlockNum(block.Header.Number)
	if err != nil {
		return nil, nil, nil, err
	}

	return block, meta, txs, nil
}

func (ps *LevelStorage) LatestBlock() (*chain.Block, error) {
	key := blockPrefixKey(latestKey)

	exists, err := ps.db.Has(key, nil)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	topKey, err := ps.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	data, err := ps.db.Get(topKey, nil)
	if err != nil {
		return nil, err
	}

	var blk chain.Block
	err = rlp.DecodeBytes(data, &blk)

	if err != nil {
		return nil, err
	}

	return &blk, nil
}

func (ps *LevelStorage) PackageBlock(txs []chain.Transaction) (*chain.BlockResult, error) {
	return ps.doPackageBlock(txs)
}

// Deposit
func (ps *LevelStorage) LastDepositPoll() (uint64, error) {
	res, err := ps.getMostRecentEventIdx(latestDepositIdxKey)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return 0, nil
		}

		return 0, err
	}

	return res, nil
}

func (ps *LevelStorage) SaveDepositPoll(idx uint64) error {
	return ps.saveEventIdx(latestDepositIdxKey, idx)
}

// Exits
func (ps *LevelStorage) LastTxExitPoll() (uint64, error) {
	return ps.getMostRecentEventIdx(lastTxExitPollKey)
}

func (ps *LevelStorage) SaveTxExitPoll(idx uint64) error {
	return ps.saveEventIdx(lastTxExitPollKey, idx)
}

func (ps *LevelStorage) LastDepositExitPoll() (uint64, error) {
	return ps.getMostRecentEventIdx(latestDepExitIdxKey)
}

func (ps *LevelStorage) SaveDepositExitPoll(idx uint64) error {
	return ps.saveEventIdx(latestDepExitIdxKey, idx)
}

func (ps *LevelStorage) SaveLastSubmittedBlock(num uint64) error {
	return ps.saveEventIdx(lastSubmittedBlockKey, num)
}

func (ps *LevelStorage) LastSubmittedBlock() (uint64, error) {
	return ps.getMostRecentEventIdx(lastSubmittedBlockKey)
}

func (ps *LevelStorage) saveEventIdx(eventKey string, idx uint64) error {
	key := prefixKey(eventKey)
	b := uint64ToBytes(idx)
	return ps.db.Put(key, b, nil)
}

func (ps *LevelStorage) getMostRecentEventIdx(eventKey string) (uint64, error) {
	key := prefixKey(eventKey)
	b, err := ps.db.Get(key, nil)
	if err == leveldb.ErrNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return bytesToUint64(b), nil
}
