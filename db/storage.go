package db

import (
    "fmt"
    "log"
    "math/big"
    "strconv"
    "sync"
    "sync/atomic"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/rlp"
    "github.com/kyokan/plasma/chain"
    "github.com/kyokan/plasma/eth"
    "github.com/kyokan/plasma/util"
    "github.com/pkg/errors"
    "github.com/syndtr/goleveldb/leveldb"
    "github.com/syndtr/goleveldb/leveldb/comparer"
    "github.com/syndtr/goleveldb/leveldb/memdb"
    levelutil "github.com/syndtr/goleveldb/leveldb/util"
)


type PlasmaStorage interface {

    isTransactionValid(tx chain.Transaction) ([]*chain.Transaction, error)
    StoreTransaction(tx chain.Transaction) error
    ProcessDeposit(tx chain.Transaction) (prev, deposit *util.MerkleTree, err error)
    FindTransactionsByBlockNum(blkNum uint64) ([]chain.Transaction, error)
    FindTransactionByBlockNumTxIdx(blkNum uint64, txIdx uint32) (*chain.Transaction, error)

    Balance(addr *common.Address) (*big.Int, error)
    SpendableTxs(addr *common.Address) ([]chain.Transaction, error)
    UTXOs(addr *common.Address) ([]chain.Transaction, error)

    BlockAtHeight(num uint64) (*chain.Block, error)
    LatestBlock() (*chain.Block, error)
    PackageCurrentBlock() (*util.MerkleTree, error)
    SaveBlock(*chain.Block) error
    //CreateGenesisBlock() (*chain.Block, error)

    LastDepositEventIdx() (uint64, error)
    SaveDepositEventIdx(idx uint64) error

    LastExitEventIdx() (uint64, error)
    SaveExitEventIdx(idx uint64) error

    GetInvalidBlock(blkHash util.Hash) (*chain.Block, error)
    SaveInvalidBlock(blk *chain.Block) error
}

type noopLock struct {
}

func (l noopLock) Lock() {
}

func (l noopLock) Unlock() {
}


type Storage struct {
    sync.RWMutex

    DB            *leveldb.DB
    MemoryDB      *memdb.DB
    BlockSize     uint32
    CurrentBlock  uint64
    PrevBlockHash util.Hash
    CurrentTxIdx  uint32
    Transactions  []chain.Transaction
}

func NewStorage(db *leveldb.DB, client eth.Client) PlasmaStorage {
    result := Storage{
        DB: db,
        MemoryDB: memdb.New(comparer.DefaultComparer, int(9 * blockSize)),
        BlockSize: blockSize,
        CurrentBlock: 1,
        CurrentTxIdx: 0,
        Transactions: make([]chain.Transaction, blockSize),
    }
    lastBlock, err := result.LatestBlock()
    if err != nil {
        log.Panic("Failed to get last block:", err)
    }

    if lastBlock == nil {
        merkle, err := result.createGenesisBlock()
        if err != nil {
            log.Panic("Failed to get last block:", err)
        }
        if client != nil {
            client.SubmitBlock(merkle)
        }
    } else {
        result.PrevBlockHash = lastBlock.BlockHash
        result.CurrentBlock  = lastBlock.Header.Number
    }

    return &result
}

func (ps *Storage) Put(key, value []byte) {
    ps.MemoryDB.Put(key, value)
}

func (ps *Storage) Delete(key []byte)  {
    ps.MemoryDB.Delete(key)
}

func (ps *Storage) findPreviousTx(tx *chain.Transaction, inputIdx uint8) (*chain.Transaction, error) {
    var input *chain.Input

    if inputIdx != 0 && inputIdx != 1 {
        panic("inputIdx must be 0 or 1")
    }

    if inputIdx == 0 {
        input = tx.Input0
    } else {
        input = tx.Input1
    }

    prevTx, err := ps.FindTransactionByBlockNumTxIdx(input.BlkNum, input.TxIdx)

    if err != nil {
        return nil, err
    }

    return prevTx, nil
}

func (ps *Storage) doStoreTransaction(tx chain.Transaction, lock sync.Locker) error {
    prevTxs, err := ps.isTransactionValid(tx)
    if err != nil {
        return err
    }
    lock.Lock()
    tx.TxIdx = atomic.AddUint32(&ps.CurrentTxIdx, 1) - 1
    tx.BlkNum = ps.CurrentBlock
    lock.Unlock()

    ps.Transactions[tx.TxIdx] = tx

    txEnc, err := rlp.EncodeToBytes(&tx)

    if err != nil {
        return err
    }

    hash := tx.Hash()
    hexHash := common.ToHex(hash)
    hashKey := txPrefixKey("hash", hexHash)

    batch := new(leveldb.Batch)
    batch.Put(hashKey, txEnc)
    batch.Put(blkNumHashkey(tx.BlkNum, hexHash), txEnc)
    batch.Put(blkNumTxIdxKey(tx.BlkNum, tx.TxIdx), txEnc)

    empty := []byte{}

    // Recording spends
    if tx.Input0.IsZeroInput() == false {
        input := tx.InputAt(0)
        outputOwner := prevTxs[0].OutputAt(input.OutIdx).NewOwner
        // TODO: Validate that signature matches
        batch.Put(spend(&outputOwner, tx.Input0), empty)
    }
    if tx.Input1.IsZeroInput() == false {
        input := tx.InputAt(0)
        outputOwner := prevTxs[1].OutputAt(input.OutIdx).NewOwner
        // TODO: Validate that signature matches
        batch.Put(spend(&outputOwner, tx.Input1), empty)
    }

    // Recording earns
    if tx.Output0.IsZeroOutput() == false {
        output := tx.OutputAt(0)
        batch.Put(earn(&output.NewOwner, tx, 0), empty)
    }
    if tx.Output1.IsZeroOutput() == false {
        output := tx.OutputAt(1)
        batch.Put(earn(&output.NewOwner, tx, 1), empty)
    }

    batch.Replay(ps)
    return nil
}

func (ps *Storage) doPackageBlock(height uint64, locker sync.Locker) (*util.MerkleTree, error) {
    // Lock for writing
    locker.Lock()
    if height != ps.CurrentBlock { // make sure we're not packaging same block twice
        locker.Unlock()
        return nil, nil
    }
    if ps.CurrentTxIdx == 0 {
        locker.Unlock()
        return nil, nil
    }
    // The batch will act as in-memory buffer
    batch := new(leveldb.Batch)
    blkNum := ps.CurrentBlock
    ps.CurrentBlock = blkNum + 1
    transactions := ps.Transactions[0:ps.CurrentTxIdx]
    ps.CurrentTxIdx = 0
    ps.Transactions = make([]chain.Transaction, blockSize)

    hashables := make([]util.Hashable, 0, len(transactions))

    for _, tx := range transactions {
        hashables = append(hashables, &tx)
    }

    merkle := util.TreeFromItems(hashables)

    rlpMerkle := rlpMerkleTree(transactions)

    header := chain.BlockHeader{
        MerkleRoot:    merkle.Root.Hash,
        RLPMerkleRoot: rlpMerkle.Root.Hash,
        PrevHash:      ps.PrevBlockHash,
        Number:        blkNum,
    }

    block := chain.Block{
        Header:    &header,
        BlockHash: header.Hash(),
    }
    ps.PrevBlockHash = block.BlockHash

    enc, err := rlp.EncodeToBytes(merkle.Root)
    if err != nil {
        locker.Unlock()
        return nil, err
    }
    batch.Put(merklePrefixKey(common.ToHex(merkle.Root.Hash)), enc)

    enc, err = rlp.EncodeToBytes(block)
    if err != nil {
        locker.Unlock()
        return nil, err
    }
    key := blockPrefixKey(common.ToHex(block.BlockHash))
    batch.Put(key, enc)
    batch.Put(blockPrefixKey(latestKey), key)
    batch.Put(blockNumKey(block.Header.Number), key)

    memIter := ps.MemoryDB.NewIterator(nil)
    for memIter.Next() {
        batch.Put(memIter.Key(), memIter.Value())
    }
    ps.MemoryDB.Reset()
    locker.Unlock()
    return &rlpMerkle, ps.DB.Write(batch, nil)
}

// TODO: return array for previous transactions
func (ps *Storage) isTransactionValid(tx chain.Transaction) ([]*chain.Transaction, error) {
    if tx.IsDeposit() {
        return []*chain.Transaction{chain.ZeroTransaction()}, nil
    }

    if tx.IsZeroTransaction() { // This may be genesis
        if ps.CurrentBlock == 1 && ps.CurrentTxIdx == 0 {
            return []*chain.Transaction{&tx}, nil
        } else {
            return nil, errors.New("Failed to add an empty transaction")
        }
    }

    result := make([]*chain.Transaction, 0, 2)
    spendKeys := make([][]byte, 0, 2)
    prevTx, err := ps.findPreviousTx(&tx, 0)
    if err != nil {
        return nil, err
    }
    if prevTx == nil {
        return nil, errors.New("Couldn't find previous transaction")
    }
    result = append(result, prevTx)
    spendKeys = append(spendKeys, spend(&prevTx.OutputAt(tx.Input0.OutIdx).NewOwner, tx.Input0))
    if tx.Input1.IsZeroInput() == false {
        prevTx, err := ps.findPreviousTx(&tx, 1)
        if err != nil {
            return nil, err
        }
        if prevTx == nil {
            return nil, errors.New("Couldn't find previous transaction")
        }
        result = append(result, prevTx)
        spendKeys = append(spendKeys, spend(&prevTx.OutputAt(tx.Input1.OutIdx).NewOwner, tx.Input1))
    }
    for i, spendKey := range spendKeys {
        if ps.MemoryDB != nil {
            found := ps.MemoryDB.Contains(spendKey)
            if found == true {
                return nil, errors.New(fmt.Sprintf("Input %d was already spent", i))
            }
        }
        found, err := ps.DB.Has(spendKey, nil)
        if err != nil {
            return nil, err
        }
        if found == true {
            return nil, errors.New(fmt.Sprintf("Input %d was already spent", i))
        }
    }
    // TODO: Validate amount
    // TODO: Validate signatures
    return result, nil
}

func (ps *Storage) StoreTransaction(tx chain.Transaction) error {
    return ps.doStoreTransaction(tx, ps.RLocker())
}

func (ps *Storage) ProcessDeposit(tx chain.Transaction) (prev, deposit *util.MerkleTree, err error) {
    ps.Lock()
    defer ps.Unlock()

    prevBlk, err := ps.doPackageBlock(ps.CurrentBlock, noopLock{})
    if err != nil {
        return nil, nil, err
    }
    ps.doStoreTransaction(tx, noopLock{})
    depositBlk, err := ps.doPackageBlock(ps.CurrentBlock, noopLock{})
    return prevBlk, depositBlk, err
}

func (ps *Storage) FindTransactionsByBlockNum(blkNum uint64) ([]chain.Transaction, error) {
    if blkNum >= ps.CurrentBlock {
        return []chain.Transaction{}, nil
    }
    ps.RLock()
    defer ps.RUnlock()

    if blkNum == ps.CurrentBlock {
        return ps.Transactions, nil
    }

    var buffer []chain.Transaction

    // Construct partial prefix that matches all transactions for the block
    prefix := txPrefixKey("blkNum", strconv.FormatUint(blkNum, 10), "txIdx")
    prefix = append(prefix, ':', ':')

    iter := ps.DB.NewIterator(levelutil.BytesPrefix(prefix), nil)
    defer iter.Release()

    for iter.Next() {
        var tx chain.Transaction
        // Extract transaction index
        // prefix looks like "tx::blkNum::1::txIdx::"
        // key looks like    "tx::blkNum::1::txIdx::20"
        idx := string(iter.Key()[len(prefix):])
        txIdx, err := strconv.ParseUint(idx, 10, 64)
        if err != nil {
            return nil, err
        }
        err = rlp.DecodeBytes(iter.Value(), &tx)
        if err != nil {
            return nil, err
        }
        // RLP encoding for tranctions doesn't contain TxIdx or BlkNum
        tx.TxIdx = uint32(txIdx)
        tx.BlkNum = blkNum
        buffer = append(buffer, tx)
    }

    txs := make([]chain.Transaction, len(buffer))
    // TODO: Do transactions have to be in index order?
    for _, tx := range buffer {
        txs[tx.TxIdx] = tx
    }

    return txs, nil
}

func (ps *Storage) FindTransactionByBlockNumTxIdx(blkNum uint64, txIdx uint32) (*chain.Transaction, error) {
    ps.RLock()
    defer ps.RUnlock()
    if blkNum == ps.CurrentBlock {
        return &ps.Transactions[txIdx], nil
    }
    key := blkNumTxIdxKey(blkNum, txIdx)
    exists, err := ps.DB.Has(key, nil)

    if err != nil {
        return nil, err
    }

    if !exists {
        return nil, nil
    }

    data, err := ps.DB.Get(key, nil)

    if err != nil {
        return nil, err
    }

    tx := chain.Transaction{}
    err = rlp.DecodeBytes(data, &tx)
    if err != nil {
        return nil, err
    }
    tx.BlkNum = blkNum
    tx.TxIdx  = txIdx

    return &tx, nil
}
// Address
func (ps *Storage) Balance(addr *common.Address) (*big.Int, error) {
    txs, err := ps.SpendableTxs(addr)

    if err != nil {
        return nil, err
    }

    total := big.NewInt(0)

    for _, tx := range txs {
        total = total.Add(total, extractAmount(&tx, addr))
    }

    return total, nil
}

func (ps *Storage) SpendableTxs(addr *common.Address) ([]chain.Transaction, error) {
    earnPrefix := earnPrefixKey(addr)
    spendPrefix := spendPrefixKey(addr)

    earnMap := make(map[string]uint8)
    spendMap := make(map[string]uint8)

    ps.RLock()

    memEarnIterator := ps.MemoryDB.NewIterator(levelutil.BytesPrefix(earnPrefix))
    defer memEarnIterator.Release()
    for memEarnIterator.Next() {
        earnKey := memEarnIterator.Key()
        lookupKey := string(earnKey[len(earnKeyPrefix) + len(keyPartsSeparator):])
        earnMap[lookupKey] = 1
    }

    memSpendIter := ps.MemoryDB.NewIterator(levelutil.BytesPrefix(spendPrefix))
    defer memSpendIter.Release()
    for memSpendIter.Next() {
        spendKey := memEarnIterator.Key()
        lookupKey := string(spendKey[len(earnKeyPrefix) + len(keyPartsSeparator):])
        spendMap[lookupKey] = 1
    }

    ps.RUnlock()

    earnIter := ps.DB.NewIterator(levelutil.BytesPrefix(earnPrefix), nil)
    defer earnIter.Release()

    for earnIter.Next() {
        earnKey := earnIter.Key()
        lookupKey := string(earnKey[len(earnKeyPrefix) + len(keyPartsSeparator):])
        earnMap[lookupKey] = 1
    }


    spendIter := ps.DB.NewIterator(levelutil.BytesPrefix(spendPrefix), nil)
    defer spendIter.Release()

    for spendIter.Next() {
        spendKey := spendIter.Key()
        lookupKey := string(spendKey[len(earnKeyPrefix) + len(keyPartsSeparator):])
        spendMap[lookupKey] = 1
    }

    for k, _ := range spendMap {
        delete(earnMap, k)
    }

    var ret []chain.Transaction
    for key := range earnMap {
        _, blkNum, txIdx, _, err := parseSuffix([]byte(key))
        if err != nil {
            return nil, err
        }
        tx, err := ps.FindTransactionByBlockNumTxIdx(blkNum, txIdx)

        if err != nil {
            return nil, err
        }

        ret = append(ret, *tx)
    }

    return ret, nil
}

func (ps *Storage) UTXOs(addr *common.Address) ([]chain.Transaction, error) {
    txs, err := ps.SpendableTxs(addr)

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
// Block
func (ps *Storage) BlockAtHeight(num uint64) (*chain.Block, error) {
    key, err := ps.DB.Get(blockNumKey(num), nil)
    if err != nil {
        return nil, err
    }
    data, err := ps.DB.Get(key, nil)
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

func (ps *Storage) LatestBlock() (*chain.Block, error) {
    key := blockPrefixKey(latestKey)

    exists, err := ps.DB.Has(key, nil)

    if err != nil {
        return nil, err
    }

    if !exists {
        return nil, nil
    }

    topKey, err := ps.DB.Get(key, nil)
    if err != nil {
        return nil, err
    }
    data, err := ps.DB.Get(topKey, nil)
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

func (ps *Storage) PackageCurrentBlock() (*util.MerkleTree, error) {
    height := atomic.LoadUint64(&ps.CurrentBlock)
    return ps.doPackageBlock(height, ps)
}

func (ps *Storage) SaveBlock(blk *chain.Block) error {
    enc, err := rlp.EncodeToBytes(blk)
    if err != nil {
        return err
    }
    ps.Lock()
    defer ps.Unlock()

    key := blockPrefixKey(common.ToHex(blk.BlockHash))
    batch := new(leveldb.Batch)
    batch.Put(key, enc)
    batch.Put(blockPrefixKey(latestKey), key)
    batch.Put(blockNumKey(blk.Header.Number), key)

    ps.CurrentBlock = blk.Header.Number + 1

    return ps.DB.Write(batch, nil)
}

func (ps *Storage) createGenesisBlock() (*util.MerkleTree, error) {
    tx := chain.Transaction{
        Input0:  chain.ZeroInput(),
        Input1:  chain.ZeroInput(),
        Sig0:    []byte{},
        Sig1:    []byte{},
        Output0: chain.ZeroOutput(),
        Output1: chain.ZeroOutput(),
        Fee:     new(big.Int),
    }
    locker := noopLock{}
    err := ps.doStoreTransaction(tx, locker)
    if err != nil {
        return nil, err
    }
    return ps.doPackageBlock(1, locker)
}
// Deposit
func (ps *Storage) LastDepositEventIdx() (uint64, error) {
    key := prefixKey(latestDepositIdxKey)
    b, err := ps.DB.Get(key, nil)
    if err != nil {
        return 0, err
    }
    return bytesToUint64(b), nil
}

func (ps *Storage) SaveDepositEventIdx(idx uint64) error {
    key := prefixKey(latestDepositIdxKey)
    b := uint64ToBytes(idx)
    return ps.DB.Put(key, b, nil)
}

// Exit
func (ps *Storage) LastExitEventIdx() (uint64, error) {
    key := prefixKey(latestExitIdxKey)
    b, err := ps.DB.Get(key, nil)
    if err != nil {
        return 0, err
    }
    return bytesToUint64(b), nil
}

func (ps *Storage) SaveExitEventIdx(idx uint64) error {
    key := prefixKey(latestExitIdxKey)
    b := uint64ToBytes(idx)
    return ps.DB.Put(key, b, nil)
}

// Invalid block
func (ps *Storage) GetInvalidBlock(blkHash util.Hash) (*chain.Block, error) {
    key := invalidPrefixKey(common.ToHex(blkHash))

    data, err := ps.DB.Get(key, nil)
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

func (ps *Storage) SaveInvalidBlock(blk *chain.Block) error {
    enc, err := rlp.EncodeToBytes(blk)
    if err != nil {
        return err
    }

    key := invalidPrefixKey(common.ToHex(blk.BlockHash))
    return ps.DB.Put(key, enc, nil)
}

