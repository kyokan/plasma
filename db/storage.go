package db

import (
    "fmt"
    "github.com/kyokan/plasma/merkle"
    "github.com/syndtr/goleveldb/leveldb/iterator"
    "github.com/syndtr/goleveldb/leveldb/opt"
    "log"
    "math/big"
    "strconv"
    "strings"
    "sync"
    "sync/atomic"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/common/hexutil"
    "github.com/ethereum/go-ethereum/rlp"
    "github.com/kyokan/plasma/chain"
    "github.com/kyokan/plasma/eth"
    "github.com/kyokan/plasma/util"
    "github.com/pkg/errors"
    "github.com/syndtr/goleveldb/leveldb"
    levelutil "github.com/syndtr/goleveldb/leveldb/util"
    "time"
)

const FeeTxIdx = 65535 // 2^16 - 1

type BlockResult struct {
    MerkleRoot         util.Hash
    NumberTransactions *big.Int
    BlockFees          *big.Int
    BlockNumber        *big.Int
}

type PlasmaStorage interface {
    StoreTransaction(tx chain.ConfirmedTransaction) (*chain.ConfirmedTransaction, error)
    ProcessDeposit(tx chain.Transaction) (prev, deposit *BlockResult, err error)
    FindTransactionsByBlockNum(blkNum uint64) ([]chain.ConfirmedTransaction, error)
    FindTransactionByBlockNumTxIdx(blkNum, txIdx *big.Int) (*chain.ConfirmedTransaction, error)

    Balance(addr *common.Address) (*big.Int, error)
    SpendableTxs(addr *common.Address) ([]chain.ConfirmedTransaction, error)
    UTXOs(addr *common.Address) ([]chain.ConfirmedTransaction, error)

    BlockAtHeight(num uint64) (*chain.Block, error)
    BlockMetaAtHeight(num uint64) (*chain.BlockMetadata, error)
    LatestBlock() (*chain.Block, error)
    PackageCurrentBlock() (result *BlockResult, err error)
    SaveBlock(*chain.Block) error

    LastDepositEventIdx() (uint64, error)
    SaveDepositEventIdx(idx uint64) error

    LastTransactionExitEventIdx() (uint64, error)
    SaveTransactionExitEventIdx(idx uint64) error

    LastDepositExitEventIdx() (uint64, error)
    SaveDepositExitEventIdx(idx uint64) error

    GetInvalidBlock(blkHash util.Hash) (*chain.Block, error)
    SaveInvalidBlock(blk *chain.Block) error

    MarkExitsAsSpent([]chain.Input) error
}

type noopLock struct {
}

func (l noopLock) Lock() {
}

func (l noopLock) Unlock() {
}

type Storage struct {
    sync.RWMutex

    DB               *leveldb.DB
    CurrentBlock     uint64
    PrevBlockHash    util.Hash
    CurrentTxIdx     int64
    CurrentBlockFees *big.Int

    Transactions     []merkle.DualHashable
}

func NewStorage(db *leveldb.DB, client eth.Client) PlasmaStorage {
    var err error
    result := Storage{
        DB: db,
        CurrentBlock: 1,
        CurrentTxIdx: 0,
        CurrentBlockFees: big.NewInt(0),
        Transactions: make([]merkle.DualHashable, 0, 1024),
    }

    if err != nil {
        log.Panic("Failed to create merkle tree:", err)
    }
    lastBlock, err := result.LatestBlock()
    if err != nil {
        log.Panic("Failed to get last block:", err)
    }

    if lastBlock == nil {
        genesisBlockResult, err := result.createGenesisBlock()
        if err != nil {
            log.Panic("Failed to get last block:", err)
        }
        if client != nil {
            client.SubmitBlock(genesisBlockResult.MerkleRoot, genesisBlockResult.NumberTransactions, genesisBlockResult.BlockFees, genesisBlockResult.BlockNumber)
        }
    } else {
        result.PrevBlockHash = lastBlock.BlockHash
        result.CurrentBlock  = 1 + lastBlock.Header.Number
    }

    return &result
}

func (ps *Storage) Put(key, value []byte) {
    ps.DB.Put(key, value, &opt.WriteOptions{ Sync: true})
}

func (ps *Storage) Delete(key []byte)  {
    ps.DB.Delete(key, &opt.WriteOptions{ Sync: true})
}

func (ps *Storage) findPreviousTx(tx *chain.ConfirmedTransaction, inputIdx uint8) (*chain.ConfirmedTransaction, util.Hash, error) {
    var input *chain.Input

    if inputIdx != 0 && inputIdx != 1 {
        panic("inputIdx must be 0 or 1")
    }

    if inputIdx == 0 {
        input = tx.Transaction.Input0
    } else {
        input = tx.Transaction.Input1
    }

    return ps.findTransactionByBlockNumTxIdx(input.BlkNum, input.TxIdx, noopLock{})
}

func (ps *Storage) doStoreTransaction(confirmed chain.ConfirmedTransaction, lock sync.Locker) (*chain.ConfirmedTransaction, error) {
    lock.Lock()
    defer lock.Unlock()

    prevTxs, err := ps.isTransactionValid(confirmed)
    if err != nil {
        return nil, err
    }
    return ps.saveTransaction(confirmed, prevTxs)
}

func (ps *Storage) saveTransaction(confirmed chain.ConfirmedTransaction, prevTxs []*chain.ConfirmedTransaction) (*chain.ConfirmedTransaction, error) {
    confirmed.Transaction.TxIdx = big.NewInt(atomic.AddInt64(&ps.CurrentTxIdx, 1) - 1)
    confirmed.Transaction.BlkNum = big.NewInt(int64(ps.CurrentBlock))

    ps.CurrentBlockFees = big.NewInt(0).Add(ps.CurrentBlockFees, confirmed.Transaction.Fee)

    txEnc, err := rlp.EncodeToBytes(&confirmed)
    if err != nil {
        return nil, err
    }

    hash := confirmed.RLPHash(util.Sha256)
    hexHash := hexutil.Encode(hash)
    hashKey := txPrefixKey("hash", hexHash)

    batch := new(leveldb.Batch)
    batch.Put(hashKey, txEnc)
    batch.Put(blkNumHashkey(confirmed.Transaction.BlkNum, hexHash), txEnc)
    batch.Put(blkNumTxIdxKey(confirmed.Transaction.BlkNum, confirmed.Transaction.TxIdx), txEnc)
    batch.Put(blockFeesKey(ps.CurrentBlock), ps.CurrentBlockFees.Bytes())

    var empty []byte

    // Recording spends
    if confirmed.Transaction.Input0.IsZeroInput() == false {
        input := confirmed.Transaction.InputAt(0)
        prevOutput := prevTxs[0].Transaction.OutputAt(input.OutIdx)
        outputOwner := prevOutput.Owner
        if confirmed.Transaction.Output0.IsExit() {
            batch.Put(spendExit(&outputOwner, confirmed.Transaction.Input0), empty)
        } else {
            batch.Put(spend(&outputOwner, confirmed.Transaction.Input0), empty)
        }
    }
    if confirmed.Transaction.Input1.IsZeroInput() == false {
        input := confirmed.Transaction.InputAt(1)
        outputOwner := prevTxs[1].Transaction.OutputAt(input.OutIdx).Owner
        batch.Put(spend(&outputOwner, confirmed.Transaction.Input1), empty)
    }

    // Recording earns
    if confirmed.Transaction.Output0.IsZeroOutput() == false {
        if confirmed.Transaction.Output0.IsDeposit() { // Only first output can be a deposit
            batch.Put(depositKey(&confirmed), txEnc)
        }
        outIdx := big.NewInt(0)
        output := confirmed.Transaction.OutputAt(outIdx)
        batch.Put(earn(&output.Owner, confirmed, outIdx), empty)
    }
    if confirmed.Transaction.Output1.IsZeroOutput() == false {
        outIdx := big.NewInt(1)
        output := confirmed.Transaction.OutputAt(outIdx)
        batch.Put(earn(&output.Owner, confirmed, outIdx), empty)
    }

    ps.Transactions = append(ps.Transactions, &confirmed)
    return &confirmed, batch.Replay(ps)
}

func (ps *Storage) MarkExitsAsSpent(inputs []chain.Input) error {
    ps.Lock()
    defer ps.Unlock()
    for _, input := range inputs {
        if input.TxIdx.Cmp(big.NewInt(FeeTxIdx)) == 0 { // fee exit
            blkNum := input.BlkNum.Uint64()
            err := ps.DB.Put(blockFeesExitKey(blkNum), input.Denom.Bytes(), nil)
            if err != nil {
                log.Printf("Error: Failed to exit fees for block %v", blkNum)
            }
            continue
        }
    	confirmed := chain.ConfirmedTransaction{
    		Transaction: *chain.ZeroTransaction(),
		}
        confirmed.Transaction.Input0 = &input
        confirmed.Transaction.Output0 = chain.ExitOutput()
        var prevTx *chain.ConfirmedTransaction
        var err error
        if input.IsDeposit() { // this is a deposit exit
            prevTx, _, err = ps.findTransactionByDepositNonce(input.DepositNonce, noopLock{})
        } else {
            prevTx, _, err = ps.findTransactionByBlockNumTxIdx(input.BlkNum, input.TxIdx, noopLock{})
        }
        if err != nil {
            log.Printf("Failed to find previous transaction(s) for exit: %s", err)
            continue
        }

        ps.saveTransaction(confirmed, []*chain.ConfirmedTransaction{prevTx})
    }

    return nil
}

func (ps *Storage) doPackageBlock(height uint64, locker sync.Locker) (*BlockResult, error) {
    // Lock for writing
    locker.Lock()
    defer locker.Unlock()
    if height != ps.CurrentBlock { // make sure we're not packaging same block twice
        return nil, nil
    }
    if ps.CurrentTxIdx == 0 {
        return nil, nil
    }
    atomic.AddUint64(&ps.CurrentBlock, 1)
    blkNum := ps.CurrentBlock - 1
    // The batch will act as in-memory buffer
    batch := new(leveldb.Batch)
    numberOfTransactions := ps.CurrentTxIdx
    currentFees := big.NewInt(ps.CurrentBlockFees.Int64())

    ps.CurrentBlockFees = big.NewInt(0)
    ps.CurrentTxIdx = 0

    rlpMerklepHash, merkleHash := merkle.GetMerkleRoot(ps.Transactions)

    header := chain.BlockHeader{
        MerkleRoot:    merkleHash,
        RLPMerkleRoot: rlpMerklepHash,
        PrevHash:      ps.PrevBlockHash,
        Number:        blkNum,
    }

    block := chain.Block{
        Header:    &header,
        BlockHash: header.Hash(),
    }
    ps.PrevBlockHash = block.BlockHash

    enc, err := rlp.EncodeToBytes(rlpMerklepHash)
    if err != nil {
       return nil, err
    }
    batch.Put(merklePrefixKey(hexutil.Encode(merkleHash)), enc)

    enc, err = rlp.EncodeToBytes(block)
    if err != nil {
        return nil, err
    }
    key := blockPrefixKey(hexutil.Encode(block.BlockHash))
    batch.Put(key, enc)
    batch.Put(blockPrefixKey(latestKey), key)
    batch.Put(blockNumKey(block.Header.Number), key)
    enc, err = rlp.EncodeToBytes(chain.BlockMetadata{
        CreatedAt: uint64(time.Now().Unix()),
    })
    if err != nil {
        return nil, err
    }
    batch.Put(blockMetaPrefixKey(block.Header.Number), enc)

    return &BlockResult{
        MerkleRoot: rlpMerklepHash,
        NumberTransactions: big.NewInt(numberOfTransactions),
        BlockFees: currentFees,
        BlockNumber: big.NewInt(int64(block.Header.Number)),
    }, ps.DB.Write(batch, nil)
}

func (ps *Storage) isTransactionValid(confirmed chain.ConfirmedTransaction) ([]*chain.ConfirmedTransaction, error) {
	empty := []*chain.ConfirmedTransaction{&chain.ConfirmedTransaction{Transaction: *chain.ZeroTransaction(), }}
	if confirmed.Transaction.IsDeposit() {
        return empty, nil
    }

    if confirmed.Transaction.IsZeroTransaction() { // This may be genesis
        if ps.CurrentBlock == 1 && ps.CurrentTxIdx == 0 {
            return empty, nil
        } else {
            return nil, errors.New("Failed to add an empty transaction")
        }
    }
    tx := confirmed.Transaction
    result := make([]*chain.ConfirmedTransaction, 0, 2)
    spendKeys := make([][]byte, 0, 4)

    prevTx, _, err := ps.findPreviousTx(&confirmed, 0)
    if err != nil {
        return nil, err
    }
    if prevTx == nil {
        return nil, errors.New("Couldn't find transaction for input 0")
    }

    outputAddress := prevTx.Transaction.OutputAt(tx.Input0.OutIdx).Owner
    signatureHash := tx.SignatureHash()
    err = util.ValidateSignature(signatureHash, confirmed.Signatures[0][:], outputAddress)
    if err != nil {
        return nil, err
    }
    err = util.ValidateSignature(tx.Input0.SignatureHash(), tx.Sig0[:], outputAddress)
    if err != nil {
        return nil, err
    }

    result = append(result, prevTx)
    spendKeys = append(spendKeys, spend(&outputAddress, tx.Input0))
    spendKeys = append(spendKeys, spendExit(&outputAddress, tx.Input0))

    if tx.Input1.IsZeroInput() == false {
        prevTx, _, err := ps.findPreviousTx(&confirmed, 1)
        if err != nil {
            return nil, err
        }
        if prevTx == nil {
            return nil, errors.New("Couldn't find transaction for input 1")
        }
        outputAddress = prevTx.Transaction.OutputAt(tx.Input1.OutIdx).Owner
        err = util.ValidateSignature(signatureHash, confirmed.Signatures[1][:], outputAddress)
        if err != nil {
            return nil, err
        }
        err = util.ValidateSignature(tx.Input1.SignatureHash(), tx.Sig1[:], outputAddress)
        if err != nil {
            return nil, err
        }
        result = append(result, prevTx)
        spendKeys = append(spendKeys, spend(&outputAddress, tx.Input1))
        spendKeys = append(spendKeys, spendExit(&outputAddress, tx.Input1))
    }

    for _, spendKey := range spendKeys {
        found, _ := ps.DB.Has(spendKey, nil)
        if found {
            msg := fmt.Sprintf("Error: Found double spend for key %s", string(spendKey))
            log.Printf(msg)
            return nil, errors.New(msg)
        }
    }

    totalInput := result[0].Transaction.OutputAt(tx.Input0.OutIdx).Denom
    if len(result) > 1 {
        totalInput = big.NewInt(0).Add(totalInput, result[1].Transaction.OutputAt(tx.Input1.OutIdx).Denom)
    }

    totalOutput := big.NewInt(0).Add(tx.Output0.Denom, tx.Fee)
    if !tx.Output1.IsZeroOutput() {
        totalOutput = totalOutput.Add(totalOutput, tx.Output1.Denom)
    }

    if totalInput.Cmp(totalOutput) != 0 {
        return nil, errors.New("inputs and outputs do not have the same sum")
    }

    return result, nil
}

func (ps *Storage) StoreTransaction(tx chain.ConfirmedTransaction) (*chain.ConfirmedTransaction, error) {
    return ps.doStoreTransaction(tx, ps)
}

func (ps *Storage) ProcessDeposit(tx chain.Transaction) (prev, deposit *BlockResult, err error) {
    ps.Lock()
    defer ps.Unlock()

    prevBlkResult, err := ps.doPackageBlock(ps.CurrentBlock, noopLock{})
    if err != nil {
        return nil, nil, err
    }
    confirmed := chain.ConfirmedTransaction{
        Transaction: tx,
    }
    ps.doStoreTransaction(confirmed, noopLock{})
    depositBlkResult, err := ps.doPackageBlock(ps.CurrentBlock, noopLock{})
    return prevBlkResult, depositBlkResult, err
}

func (ps *Storage) FindTransactionsByBlockNum(blkNum uint64) ([]chain.ConfirmedTransaction, error) {
    if blkNum >= ps.CurrentBlock {
        return []chain.ConfirmedTransaction{}, nil
    }
    ps.RLock()
    defer ps.RUnlock()

    // Construct partial prefix that matches all transactions for the block
    prefix := txPrefixKey("blkNum", strconv.FormatUint(blkNum, 10), "txIdx")
    prefix = append(prefix, ':', ':')


    iter := ps.DB.NewIterator(levelutil.BytesPrefix(prefix), nil)
    defer iter.Release()

    return findBlockTransactions(iter, prefix, blkNum)
}

func findBlockTransactions(iter iterator.Iterator, prefix []byte, blkNum uint64) ([]chain.ConfirmedTransaction, error) {
    var buffer []chain.ConfirmedTransaction

    for iter.Next() {
        var tx chain.ConfirmedTransaction
        // Extract transaction index
        // prefix looks like "tx::blkNum::1::txIdx::"
        // key looks like    "tx::blkNum::1::txIdx::20"
        idx := string(iter.Key()[len(prefix):])
        txIdx, success := new(big.Int).SetString(idx, 10)
        if success == false {
            return nil, errors.New("Failed to parse transaction index from key")
        }
        err := rlp.DecodeBytes(iter.Value(), &tx)
        if err != nil {
            return nil, err
        }
        // RLP encoding for tranctions doesn't contain TxIdx or BlkNum
        tx.Transaction.TxIdx = txIdx
        tx.Transaction.BlkNum = big.NewInt(int64(blkNum))
        buffer = append(buffer, tx)
    }

    txs := make([]chain.ConfirmedTransaction, len(buffer))
    for _, tx := range buffer {
        txs[tx.Transaction.TxIdx.Int64()] = tx
    }

    return txs, nil
}

func (ps *Storage) findTransactionByDepositNonce(nonce *big.Int, locker sync.Locker) (*chain.ConfirmedTransaction, util.Hash, error) {
    locker.Lock()
    defer locker.Unlock()

    keyPrefix := depositPrefixKey(nonce)
    iter := ps.DB.NewIterator(levelutil.BytesPrefix(keyPrefix), nil)
    defer iter.Release()

    for iter.Next() {
        var confirmed chain.ConfirmedTransaction
        key := string(iter.Key())
        keyParts := strings.Split(key, "::")
        if len(keyParts) != 4 {
            return nil, nil, errors.New(fmt.Sprintf("Failed to parse deposit transaction position from key %s", key))
        }
        blkNum, success := new(big.Int).SetString(keyParts[2], 10)
        if success == false {
            return nil, nil, errors.New(fmt.Sprintf("Failed to parse block number from key %s", key))
        }
        txIdx, success := new(big.Int).SetString(keyParts[3], 10)
        if success == false {
            return nil, nil, errors.New(fmt.Sprintf("Failed to parse transaction index from key %s", key))
        }
        err := rlp.DecodeBytes(iter.Value(), &confirmed)
        if err != nil {
            return nil, nil, err
        }
        confirmed.Transaction.BlkNum = blkNum
        confirmed.Transaction.TxIdx = txIdx
        return &confirmed, nil, nil
    }
    return nil, nil, errors.New(fmt.Sprintf("Failed to find deposit for deposit nonce %s", nonce.String()))
}

func (ps *Storage) findTransactionByBlockNumTxIdx(blkNum, txIdx *big.Int, locker sync.Locker) (*chain.ConfirmedTransaction, util.Hash, error) {
    locker.Lock()
    defer locker.Unlock()

    key := blkNumTxIdxKey(blkNum, txIdx)
    var data []byte
    var err error

    block, err := ps.BlockAtHeight(blkNum.Uint64())
    if err != nil {
        return nil, nil, err
    }

    exists, err := ps.DB.Has(key, nil)
    if err != nil {
        return nil, nil, err
    }
    if !exists {
        return nil, nil, nil
    }
    data, err = ps.DB.Get(key, nil)
    if err != nil {
        return nil, nil, err
    }

    tx := chain.ConfirmedTransaction{}
    err = rlp.DecodeBytes(data, &tx)
    if err != nil {
        return nil, nil, err
    }
    tx.Transaction.BlkNum = blkNum
    tx.Transaction.TxIdx  = txIdx

    return &tx, block.BlockHash, nil
}

func (ps *Storage) FindTransactionByBlockNumTxIdx(blkNum, txIdx *big.Int) (*chain.ConfirmedTransaction, error) {
    tx, _, err := ps.findTransactionByBlockNumTxIdx(blkNum, txIdx, ps.RLocker())
    return tx, err
}
// Address
func (ps *Storage) Balance(addr *common.Address) (*big.Int, error) {
    txs, err := ps.SpendableTxs(addr)

    if err != nil {
        return nil, err
    }

    total := big.NewInt(0)

    for _, confirmed := range txs {
        total = total.Add(total, extractAmount(&confirmed.Transaction, addr))
    }

    return total, nil
}

func (ps *Storage) SpendableTxs(addr *common.Address) ([]chain.ConfirmedTransaction, error) {
    earnPrefix := earnPrefixKey(addr)
    spendPrefix := spendPrefixKey(addr)

    earnMap := make(map[string]uint8)
    spendMap := make(map[string]uint8)

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
        lookupKey := string(spendKey[len(spendKeyPrefix) + len(keyPartsSeparator):])
        spendMap[lookupKey] = 1
    }

    for k, _ := range spendMap {
        delete(earnMap, k)
    }

    var ret []chain.ConfirmedTransaction
    for key := range earnMap {
        _, blkNum, txIdx, _, err := parseSuffix([]byte(key))
        if err != nil {
            return nil, err
        }
        tx, _, err := ps.findTransactionByBlockNumTxIdx(blkNum, txIdx, noopLock{})

        if err != nil {
            return nil, err
        }

        ret = append(ret, *tx)
    }
    sortTransactions(ret)
    return ret, nil
}

func (ps *Storage) UTXOs(addr *common.Address) ([]chain.ConfirmedTransaction, error) {
    earnPrefix := earnPrefixKey(addr)
    earnMap := make(map[string]uint8)

    earnIter := ps.DB.NewIterator(levelutil.BytesPrefix(earnPrefix), nil)
    defer earnIter.Release()

    for earnIter.Next() {
        earnKey := earnIter.Key()
        lookupKey := string(earnKey[len(earnKeyPrefix) + len(keyPartsSeparator):])
        earnMap[lookupKey] = 1
    }

    var ret []chain.ConfirmedTransaction
    for key := range earnMap {
        _, blkNum, txIdx, _, err := parseSuffix([]byte(key))
        if err != nil {
            return nil, err
        }
        tx, _, err := ps.findTransactionByBlockNumTxIdx(blkNum, txIdx, noopLock{})

        if err != nil {
            return nil, err
        }

        ret = append(ret, *tx)
    }
    sortTransactions(ret)
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

func (ps *Storage) BlockMetaAtHeight(num uint64) (*chain.BlockMetadata, error) {
    data, err := ps.DB.Get(blockMetaPrefixKey(num), nil)
    if err != nil {
        return nil, err
    }

    var meta chain.BlockMetadata
    err = rlp.DecodeBytes(data, &meta)
    if err != nil {
        return nil, err
    }

    return &meta, nil
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

func (ps *Storage) PackageCurrentBlock() (*BlockResult, error) {
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

    key := blockPrefixKey(hexutil.Encode(blk.BlockHash))
    batch := new(leveldb.Batch)
    batch.Put(key, enc)
    batch.Put(blockPrefixKey(latestKey), key)
    batch.Put(blockNumKey(blk.Header.Number), key)

    ps.CurrentBlock = blk.Header.Number + 1

    return ps.DB.Write(batch, nil)
}

func (ps *Storage) createGenesisBlock() (*BlockResult, error) {
    confirmed := chain.ConfirmedTransaction{
        Transaction: *chain.ZeroTransaction(),
    }

    locker := noopLock{}
    _, err := ps.doStoreTransaction(confirmed, locker)
    if err != nil {
        return nil, err
    }
    return ps.doPackageBlock(1, locker)
}
// Deposit
func (ps *Storage) LastDepositEventIdx() (uint64, error) {
    return ps.getMostRecentEventIdx(latestDepositIdxKey)
}

func (ps *Storage) SaveDepositEventIdx(idx uint64) error {
    return ps.saveEventIdx(latestDepositIdxKey, idx)
}

// Exits
func (ps *Storage) LastTransactionExitEventIdx() (uint64, error) {
    return ps.getMostRecentEventIdx(latestTxExitIdxKey)
}

func (ps *Storage) SaveTransactionExitEventIdx(idx uint64) error {
    return ps.saveEventIdx(latestTxExitIdxKey, idx)
}

func (ps *Storage) LastDepositExitEventIdx() (uint64, error) {
    return ps.getMostRecentEventIdx(latestDepExitIdxKey)
}

func (ps *Storage) SaveDepositExitEventIdx(idx uint64) error {
    return ps.saveEventIdx(latestDepExitIdxKey, idx)
}

func (ps *Storage) saveEventIdx(eventKey string, idx uint64) error {
    key := prefixKey(eventKey)
    b := uint64ToBytes(idx)
    return ps.DB.Put(key, b, nil)
}

func (ps *Storage) getMostRecentEventIdx(eventKey string) (uint64, error) {
    key := prefixKey(eventKey)
    b, err := ps.DB.Get(key, nil)
    if err != nil {
        return 0, err
    }
    return bytesToUint64(b), nil
}

// Invalid block
func (ps *Storage) GetInvalidBlock(blkHash util.Hash) (*chain.Block, error) {
    key := invalidPrefixKey(hexutil.Encode(blkHash))

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

    key := invalidPrefixKey(hexutil.Encode(blk.BlockHash))
    return ps.DB.Put(key, enc, nil)
}

