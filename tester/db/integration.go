package db

import (
	"log"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/util"
	"gopkg.in/urfave/cli.v1"
)

func IntegrationTest(c *cli.Context) {
	dburl := c.GlobalString("db")

	db, storage, err := db.CreateStorage(dburl, nil)

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	blockTest(storage)
	txTest(storage)
}

func blockTest(storage db.PlasmaStorage) {
	_, err := storage.BlockAtHeight(1)

	if err != nil {
		panic(err)
	}
}

func txTest(storage db.PlasmaStorage) {
	userAddress := common.HexToAddress("2263dd78-b1de-4d26-a644-a8fa9448e51d")

	txs := []chain.Transaction{
		createTestTransaction(
			1,
			0,
			chain.ZeroInput(),
			&chain.Output{
				NewOwner: userAddress,
				Amount:   util.NewInt64(100),
			},
		),
		createTestTransaction(
			1,
			1,
			&chain.Input{
				BlkNum: 1,
				TxIdx:  0,
				OutIdx: 0,
			},
			&chain.Output{
				NewOwner: userAddress,
				Amount:   util.NewInt64(100),
			},
		),
	}

	for _, tx := range txs {
		err := storage.StoreTransaction(tx)

		if err != nil {
			panic(err)
		}
	}

	resTxs, err := storage.FindTransactionByBlockNum(1)

	if err != nil {
		panic(err)
	}

	assert(len(resTxs) == 2)
}

func createTestTransaction(
	blknum uint64,
	txId uint32,
	input0 *chain.Input,
	output0 *chain.Output,
) chain.Transaction {
	return chain.Transaction{
		Input0:  input0,
		Input1:  chain.ZeroInput(),
		Sig0:    []byte{},
		Sig1:    []byte{},
		Output0: output0,
		Output1: chain.ZeroOutput(),
		Fee:     new(big.Int),
		BlkNum:  blknum,
		TxIdx:   txId,
	}
}

func assert(result bool) {
	if !result {
		panic("Assert failed!")
	}
}

func assertEquals(a interface{}, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
