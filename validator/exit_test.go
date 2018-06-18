package validator

import (
    "math/big"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/kyokan/plasma/chain"
    "github.com/kyokan/plasma/db"
    dbMocks "github.com/kyokan/plasma/db/mocks"
    "github.com/kyokan/plasma/eth"
    plasma_rpc "github.com/kyokan/plasma/rpc"
    "github.com/kyokan/plasma/test_util"
    clientMocks "github.com/kyokan/plasma/userclient/mocks"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

func mockDB(
        transactionDao *dbMocks.TransactionDao,
        blockDao *dbMocks.BlockDao,
        merkleDao *dbMocks.MerkleDao,
        addressDao *dbMocks.AddressDao,
        depositDao *dbMocks.DepositDao,
        exitDao *dbMocks.ExitDao,
        invalidBlockDao *dbMocks.InvalidBlockDao) *db.Database {
    return &db.Database{
        TxDao:           transactionDao,
        BlockDao:        blockDao,
        MerkleDao:       merkleDao,
        AddressDao:      addressDao,
        DepositDao:      depositDao,
        ExitDao:         exitDao,
        InvalidBlockDao: invalidBlockDao,
    }
}

type exit struct {
    Owner     common.Address `json:"Owner"`
    Amount    int64          `json:"Amount"`
    BlockNum  int64          `json:"BlockNum"`
    TxIndex   int64          `json:"TxIndex"`
    OIndex    int64          `json:"OIndex"`
    StartedAt int64          `json:"StartedAt"`
}

func (e *exit) toEthExit() eth.Exit {
    result := eth.Exit{
        Amount    : big.NewInt(e.Amount),
        BlockNum  : big.NewInt(e.BlockNum),
        TxIndex   : big.NewInt(e.TxIndex),
        OIndex    : big.NewInt(e.OIndex),
        StartedAt : big.NewInt(e.StartedAt),
    }
    result.Owner = e.Owner
    return result
}

type validator_fixture struct {
    Blocks []plasma_rpc.GetBlocksResponse `json:"blocks"`
    Exit   exit                           `json:"exit"`
}

func (f validator_fixture) GetBlock(height uint64) *plasma_rpc.GetBlocksResponse {
    max := uint64(0)
    var latest *plasma_rpc.GetBlocksResponse
    for _, block := range f.Blocks {
        if max < block.Block.Header.Number {
            max = block.Block.Header.Number
            latest = &block
        }
        if block.Block.Header.Number == height {
            return &block
        }
    }
    return latest
}

func (f validator_fixture) GetLatest() (*chain.Block, error) {
    response := f.GetBlock(1000)
    return response.Block, nil
}

func Test_FindDoubleSpend(t *testing.T) {
    fixture := validator_fixture{}
    err := test_util.LoadFixture(t, &fixture)
    require.NoError(t, err)
    blockDao := new(dbMocks.BlockDao)
    blockDao.On("Latest").Return(fixture.GetBlock(1000).Block, nil)
    db := mockDB(nil, blockDao, nil, nil, nil, nil, nil)
    rootClient := new(clientMocks.RootClient)
    getBlock := func(height uint64) *plasma_rpc.GetBlocksResponse {
        return fixture.GetBlock(height)
    }
    rootClient.On("GetBlock", mock.MatchedBy(func (uint64) bool { return true })).Return(getBlock)
    FindDoubleSpend(rootClient, db, nil, fixture.Exit.toEthExit())
}
