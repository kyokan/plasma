package validator

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/node"
	"github.com/kyokan/plasma/userclient"
	"github.com/kyokan/plasma/util"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
)

// TODO: add an exit client to make from the command line.
func ExitStartedListener(rootPort int, level *db.Database, plasma *eth.PlasmaClient) {
	rootUrl := fmt.Sprintf("http://localhost:%d/rpc", rootPort)

	for {
		// TODO: change name to block number
		idx, err := level.ExitDao.LastExitEventIdx()

		if err != nil && err.Error() != "leveldb: not found" {
			panic(err)
		}

		log.Printf("Looking for exit events at block number: %d\n", idx)

		events, lastIdx := plasma.ExitStartedFilter(idx)

		if len(events) > 0 {
			count := uint64(0)

			for _, event := range events {
				count += 1

				exitId := event.ExitId

				fmt.Println("Found exit id")
				fmt.Println(exitId.Uint64())
				exit := plasma.GetExit(exitId)
				fmt.Println("**** exit found")
				fmt.Println(exit)

				txs, blockId, txId := FindDoubleSpend(rootUrl, level, plasma, exit)

				if txs != nil && txId != nil {
					fmt.Println("inputs to the challenge!")
					fmt.Println(exitId)
					fmt.Println(blockId)
					fmt.Println(txId)
					plasma.ChallengeExit(
						exitId,
						txs,
						blockId,
						txId,
					)

					time.Sleep(3 * time.Second)

					events, _ := plasma.ChallengeSuccessFilter(0)

					for _, event := range events {
						fmt.Println("success")
						fmt.Println(event.ExitId)
					}

					events2, _ := plasma.ChallengeFailureFilter(0)

					for _, event := range events2 {
						fmt.Println("failure")
						fmt.Println(event.ExitId)
					}
				}

				// TODO: also if someone exits on the plasma chain you need to
				// make sure you exit it from the root node.
				// So the root node also needs an exit listener.

				// There's a race condition where someone could try to spend
				// while an exit is happenning

				// This sort of implies that you should be validating exits
				// often, not just on notification.

				// It's not synchronized right now...
				time.Sleep(time.Second * 3)
			}

			log.Printf("Found %d exit events at from blocks %d to %d.\n", count, idx, lastIdx)

			level.ExitDao.SaveExitEventIdx(lastIdx + 1)
		} else {
			log.Printf("No exit events at block %d.\n", idx)
		}

		time.Sleep(time.Second * 10)
	}
}

// TODO: move this struct and reuse in exit_client.
type TransactionInfo struct {
	block    *chain.Block
	txs      []chain.Transaction
	blocknum *big.Int
	txindex  *big.Int
}

func FindDoubleSpend(rootUrl string, level *db.Database, plasma *eth.PlasmaClient, exit eth.Exit) ([]chain.Transaction, *big.Int, *big.Int) {
	latestBlock, err := level.BlockDao.Latest()

	if err != nil {
		panic(err)
	}

	txIdx := exit.TxIndex.Uint64()
	lastBlockHeight := latestBlock.Header.Number
	currBlockHeight := exit.BlockNum.Uint64() + 1

	response := userclient.GetBlock(rootUrl, exit.BlockNum.Uint64())

	if txIdx >= uint64(len(response.Transactions)) {
		log.Fatalf("The following exit does not exist within this block!")
	}

	exitTx := response.Transactions[exit.TxIndex.Uint64()]

	fmt.Printf("Finding spends from blocks %d to %d\n", currBlockHeight, lastBlockHeight)

	// Find possible double spends in every block
	// TODO: actually in theory it should never happen in the current block.
	// Because root node will never create and submit that block.
	// Also, how do you protect against exits happenning more than once?
	for i := currBlockHeight; i <= lastBlockHeight; i++ {
		response := userclient.GetBlock(rootUrl, i)
		currTxs := response.Transactions
		rej := node.FindMatchingInputs(&exitTx, currTxs)

		if len(rej) > 0 {
			fmt.Printf("Found %d double spends at block %d\n", len(rej), i)

			fmt.Println(exit.BlockNum)
			fmt.Println(exit.TxIndex)
			fmt.Println(exit.OIndex)
			fmt.Println(rej[0].BlkNum)
			fmt.Println(rej[0].Hash())
			fmt.Println(rej[0].Input0.BlkNum)
			fmt.Println(rej[0].Input0.TxIdx)
			fmt.Println(rej[0].Input0.OutIdx)
			fmt.Println(rej[0].Input1.BlkNum)
			fmt.Println(rej[0].Input1.TxIdx)
			fmt.Println(rej[0].Input1.OutIdx)
			fmt.Println(rej[0].Output0.NewOwner)
			fmt.Println(rej[0].Output0.Amount)
			fmt.Println(rej[0].Output1.NewOwner)
			fmt.Println(rej[0].Output1.Amount)
			// Always return the first one.
			return currTxs, util.NewUint64(i), util.NewUint32(rej[0].TxIdx)
		} else {
			fmt.Printf("Found no double spends for block %d\n", i)
		}
	}

	return nil, nil, nil
}
