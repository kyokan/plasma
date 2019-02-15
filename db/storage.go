package db

import (
	"fmt"
	"github.com/kyokan/plasma/merkle"
	"math/big"
	"strings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/util"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	levelutil "github.com/syndtr/goleveldb/leveldb/util"
	"time"
	"github.com/golang/protobuf/proto"
)

const FeeTxIdx = 65535 // 2^16 - 1

type BlockResult struct {
	MerkleRoot         util.Hash
	NumberTransactions uint32
	BlockFees          *big.Int
	BlockNumber        *big.Int
}

type PlasmaStorage interface {
	ProcessDeposit(tx chain.Transaction) (deposit *BlockResult, err error)
	FindTransactionsByBlockNum(blkNum uint64) ([]chain.ConfirmedTransaction, error)
	FindTransactionByBlockNumTxIdx(blkNum uint64, txIdx uint32) (*chain.ConfirmedTransaction, error)

	Balance(addr common.Address) (*big.Int, error)
	SpendableTxs(addr common.Address) ([]chain.ConfirmedTransaction, error)
	UTXOs(addr common.Address) ([]chain.ConfirmedTransaction, error)

	BlockAtHeight(num uint64) (*chain.Block, error)
	BlockMetaAtHeight(num uint64) (*chain.BlockMetadata, error)
	LatestBlock() (*chain.Block, error)
	PackageBlock(txs []chain.Transaction) (result *BlockResult, err error)
	ConfirmTransaction(blockNumber uint64, transactionIndex uint32, sigs [2]chain.Signature) (*chain.ConfirmedTransaction, error)
	ConfirmSigsFor(blockNumber uint64, transactionIndex uint32) ([2]chain.Signature, error)

	LastDepositPoll() (uint64, error)
	SaveDepositPoll(idx uint64) error

	LastTxExitPoll() (uint64, error)
	SaveTxExitPoll(idx uint64) error

	LastDepositExitEventIdx() (uint64, error)
	SaveDepositExitEventIdx(idx uint64) error

	MarkExitsAsSpent([]chain.Input) error

	IsDoubleSpent(tx *chain.Transaction) (bool, error)

	SaveLastSubmittedBlock(num uint64) error
	LastSubmittedBlock() (uint64, error)

	FindDoubleSpendingTransaction(blkNum uint64, txIdx uint32, outIndex uint8) (*chain.ConfirmedTransaction, error)
}

type Storage struct {
	db *leveldb.DB
}

func NewStorage(db *leveldb.DB) PlasmaStorage {
	result := Storage{
		db: db,
	}

	return &result
}

func (ps *Storage) Put(key, value []byte) {
	ps.db.Put(key, value, nil)
}

func (ps *Storage) Delete(key []byte) {
	ps.db.Delete(key, nil)
}

func (ps *Storage) findPreviousTx(tx *chain.Transaction, inputIdx uint8) (*chain.ConfirmedTransaction, error) {
	var input *chain.Input
	if inputIdx != 0 && inputIdx != 1 {
		panic("inputIdx must be 0 or 1")
	}

	if inputIdx == 0 {
		input = tx.Body.Input0
	} else {
		input = tx.Body.Input1
	}

	return ps.findTransactionByBlockNumTxIdx(input.BlockNum, input.TxIdx)
}

func (ps *Storage) saveTransaction(blockNum uint64, txIdx uint32, originalTx chain.Transaction, batch *leveldb.Batch) (*chain.ConfirmedTransaction, error) {
	tx := originalTx.Clone()
	tx.Body.TransactionIndex = txIdx
	tx.Body.BlockNumber = blockNum

	confirmed := &chain.ConfirmedTransaction{
		Transaction: tx,
	}
	confirmedEnc, err := proto.Marshal(confirmed.Proto())
	if err != nil {
		return nil, err
	}
	hash := confirmed.Hash()
	hexHash := hexutil.Encode(hash)

	batch.Put(txByHashKey(hexHash), confirmedEnc)
	batch.Put(txByBlockNumTxIdxKey(blockNum, txIdx), []byte(hexHash))

	var empty []byte

	// Recording spends
	if tx.Body.IsDeposit() {
		batch.Put(depositKey(tx.Body.Input0.DepositNonce), []byte(hexHash))
	} else {
		prevConfirmed0, err := ps.findPreviousTx(tx, 0)
		if err != nil {
			return nil, err
		}

		prevTx := prevConfirmed0.Transaction
		batch.Put(spendByTxIdxKey(prevTx.Body.BlockNumber, prevTx.Body.TransactionIndex, tx.Body.Input0.OutIdx), []byte(hexHash))
		batch.Delete(utxoKey(prevTx.Body.OutputAt(tx.Body.Input0.OutIdx).Owner, prevConfirmed0.Hash(), tx.Body.Input0.OutIdx))
		// TODO mark as exited
	}

	if !tx.Body.Input1.IsZero() {
		prevConfirmed1, err := ps.findPreviousTx(tx, 1)
		if err != nil {
			return nil, err
		}

		prevTx := prevConfirmed1.Transaction
		batch.Put(spendByTxIdxKey(prevTx.Body.BlockNumber, prevTx.Body.TransactionIndex, tx.Body.Input1.OutIdx), []byte(hexHash))
		batch.Delete(utxoKey(prevTx.Body.OutputAt(tx.Body.Input0.OutIdx).Owner, prevConfirmed1.Hash(), tx.Body.Input1.OutIdx))
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

	return confirmed, batch.Replay(ps)
}

func (ps *Storage) MarkExitsAsSpent(inputs []chain.Input) error {
	//for _, input := range inputs {
	//	if input.TransactionIndex.Cmp(big.NewInt(FeeTxIdx)) == 0 { // fee exit
	//		blkNum := input.BlockNumber.Uint64()
	//		err := ps.db.Put(blockFeesExitKey(blkNum), input.Amount.Bytes(), nil)
	//		if err != nil {
	//			log.Printf("Error: Failed to exit fees for block %v", blkNum)
	//		}
	//		continue
	//	}
	//	confirmed := chain.Body{
	//		TransactionBody: *chain.ZeroBody(),
	//	}
	//	confirmed.TransactionBody.Input0 = &input
	//	confirmed.TransactionBody.Output0 = chain.ExitOutput()
	//	var prevTx *chain.Body
	//	var err error
	//	if input.IsDeposit() { // this is a deposit exit
	//		prevTx, _, err = ps.findTransactionByDepositNonce(input.DepositNonce, noopLock{})
	//	} else {
	//		prevTx, _, err = ps.findTransactionByBlockNumTxIdx(input.BlockNumber.Uint64(), uint32(input.TransactionIndex.Uint64()), noopLock{})
	//	}
	//	if err != nil {
	//		log.Printf("Failed to find previous transaction(s) for exit: %s", err)
	//		continue
	//	}
	//
	//	//ps.saveTransaction(confirmed, []*chain.Body{prevTx})
	//}

	return nil
}

func (ps *Storage) IsDoubleSpent(tx *chain.Transaction) (bool, error) {
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
	searchKeys = append(searchKeys, spendByTxIdxKey(prevTx0Body.BlockNumber, prevTx0Body.TransactionIndex, body.Input0.OutIdx))
	if !body.Input1.IsZero() {
		prevTx1, err := ps.findPreviousTx(tx, 1)
		if err != nil {
			return false, err
		}
		prevTx1Body := prevTx1.Transaction.Body
		searchKeys = append(searchKeys, spendByTxIdxKey(prevTx1Body.BlockNumber, prevTx1Body.TransactionIndex, body.Input1.OutIdx))
	}

	for _, spendKey := range searchKeys {
		found, _ := ps.db.Has(spendKey, nil)
		if found {
			return true, nil
		}
	}

	return false, nil
}

func (ps *Storage) FindDoubleSpendingTransaction(blkNum uint64, txIdx uint32, outIndex uint8) (*chain.ConfirmedTransaction, error) {
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

func (ps *Storage) doPackageBlock(txs []chain.Transaction) (*BlockResult, error) {
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
	block := chain.Block{
		Header:    &header,
		BlockHash: header.Hash(),
	}

	enc, err := rlp.EncodeToBytes(merkleRoot)
	if err != nil {
		return nil, err
	}
	batch.Put(merklePrefixKey(hexutil.Encode(merkleRoot)), enc)

	enc, err = rlp.EncodeToBytes(block)
	if err != nil {
		return nil, err
	}
	key := blockPrefixKey(hexutil.Encode(block.BlockHash))
	batch.Put(key, enc)
	batch.Put(blockPrefixKey(latestKey), key)
	batch.Put(blockNumKey(block.Header.Number), key)

	currentFees := big.NewInt(0)
	for i, tx := range txs {
		_, err := ps.saveTransaction(blkNum, uint32(i), tx, batch)
		if err != nil {
			return nil, err
		}

		currentFees = currentFees.Add(currentFees, tx.Body.Fee)
	}
	batch.Put(blockFeesKey(blkNum), currentFees.Bytes())

	meta := &chain.BlockMetadata{
		CreatedAt:        uint64(time.Now().Unix()),
		TransactionCount: uint32(numberOfTransactions),
		Fees:             currentFees,
	}
	metaEnc, err := meta.RLP()
	if err != nil {
		return nil, err
	}
	batch.Put(blockMetaPrefixKey(block.Header.Number), metaEnc)

	err = ps.db.Write(batch, nil)
	if err != nil {
		return nil, err
	}

	return &BlockResult{
		MerkleRoot:         merkleRoot,
		NumberTransactions: uint32(numberOfTransactions),
		BlockFees:          currentFees,
		BlockNumber:        big.NewInt(int64(block.Header.Number)),
	}, nil
}

func (ps *Storage) ProcessDeposit(confirmed chain.Transaction) (deposit *BlockResult, err error) {
	if !confirmed.Body.IsDeposit() {
		return nil, errors.New("only deposit blocks are accepted")
	}

	return ps.doPackageBlock([]chain.Transaction{confirmed})
}

func (ps *Storage) FindTransactionsByBlockNum(blockNum uint64) ([]chain.ConfirmedTransaction, error) {
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

func (ps *Storage) findTransactionByDepositNonce(nonce *big.Int) (*chain.Transaction, util.Hash, error) {
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

func (ps *Storage) findTransactionByBlockNumTxIdx(blockNum uint64, txIdx uint32) (*chain.ConfirmedTransaction, error) {
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

func (ps *Storage) findTransactionByHash(hash util.Hash) (*chain.ConfirmedTransaction, error) {
	txBytes, err := ps.db.Get(txByHashKey(hexutil.Encode(hash)), nil)
	if err != nil {
		return nil, err
	}
	return chain.UnmarshalConfirmedTransactionProto(txBytes)
}

func (ps *Storage) FindTransactionByBlockNumTxIdx(blkNum uint64, txIdx uint32) (*chain.ConfirmedTransaction, error) {
	tx, err := ps.findTransactionByBlockNumTxIdx(blkNum, txIdx)
	return tx, err
}

// Address
func (ps *Storage) Balance(addr common.Address) (*big.Int, error) {
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

func (ps *Storage) SpendableTxs(addr common.Address) ([]chain.ConfirmedTransaction, error) {
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

func (ps *Storage) UTXOs(addr common.Address) ([]chain.ConfirmedTransaction, error) {
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

func (ps *Storage) ConfirmTransaction(blockNumber uint64, transactionIndex uint32, sigs [2]chain.Signature) (*chain.ConfirmedTransaction, error) {
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

func (ps *Storage) ConfirmSigsFor(blockNumber uint64, transactionIndex uint32) ([2]chain.Signature, error) {
	var sigs [2]chain.Signature
	tx, err := ps.findTransactionByBlockNumTxIdx(blockNumber, transactionIndex)
	if err != nil {
		return sigs, err
	}

	return tx.ConfirmSigs, err
}

// Block
func (ps *Storage) BlockAtHeight(num uint64) (*chain.Block, error) {
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

func (ps *Storage) BlockMetaAtHeight(num uint64) (*chain.BlockMetadata, error) {
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

func (ps *Storage) LatestBlock() (*chain.Block, error) {
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

func (ps *Storage) PackageBlock(txs []chain.Transaction) (*BlockResult, error) {
	return ps.doPackageBlock(txs)
}

// Deposit
func (ps *Storage) LastDepositPoll() (uint64, error) {
	res, err := ps.getMostRecentEventIdx(latestDepositIdxKey)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return 0, nil
		}

		return 0, err
	}

	return res, nil
}

func (ps *Storage) SaveDepositPoll(idx uint64) error {
	return ps.saveEventIdx(latestDepositIdxKey, idx)
}

// Exits
func (ps *Storage) LastTxExitPoll() (uint64, error) {
	return ps.getMostRecentEventIdx(lastTxExitPollKey)
}

func (ps *Storage) SaveTxExitPoll(idx uint64) error {
	return ps.saveEventIdx(lastTxExitPollKey, idx)
}

func (ps *Storage) LastDepositExitEventIdx() (uint64, error) {
	return ps.getMostRecentEventIdx(latestDepExitIdxKey)
}

func (ps *Storage) SaveDepositExitEventIdx(idx uint64) error {
	return ps.saveEventIdx(latestDepExitIdxKey, idx)
}

func (ps *Storage) SaveLastSubmittedBlock(num uint64) error {
	return ps.saveEventIdx(lastSubmittedBlockKey, num)
}

func (ps *Storage) LastSubmittedBlock() (uint64, error) {
	return ps.getMostRecentEventIdx(lastSubmittedBlockKey)
}

func (ps *Storage) saveEventIdx(eventKey string, idx uint64) error {
	key := prefixKey(eventKey)
	b := uint64ToBytes(idx)
	return ps.db.Put(key, b, nil)
}

func (ps *Storage) getMostRecentEventIdx(eventKey string) (uint64, error) {
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
