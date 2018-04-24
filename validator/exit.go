package validator

import (
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

func ExitStartedListener(rootUrl string, level *db.Database, plasma *eth.PlasmaClient) {
	for {
		idx, err := level.ExitDao.LastExitEventIdx()

		if err != nil && err.Error() != "leveldb: not found" {
			log.Fatalf("Failed to get last exit event idx: %v", err)
		}

		log.Printf("Looking for exit events at block number: %d\n", idx)

		events, lastIdx := plasma.ExitStartedFilter(idx)

		if len(events) > 0 {
			count := uint64(0)

			for _, event := range events {
				count += 1

				exitId := event.ExitId

				exit := plasma.GetExit(exitId)

				txs, blockId, txId := FindDoubleSpend(rootUrl, level, plasma, exit)

				if txs != nil && txId != nil {
					plasma.ChallengeExit(
						exitId,
						txs,
						blockId,
						txId,
					)

					time.Sleep(3 * time.Second)

					events, _ := plasma.ChallengeSuccessFilter(0)

					for _, event := range events {
						log.Printf("challenge success: %v", event.ExitId)
					}

					events2, _ := plasma.ChallengeFailureFilter(0)

					for _, event := range events2 {
						log.Printf("challenge failure: %v", event.ExitId)
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

func FindDoubleSpend(rootUrl string, level *db.Database, plasma *eth.PlasmaClient, exit eth.Exit) ([]chain.Transaction, *big.Int, *big.Int) {
	latestBlock, err := level.BlockDao.Latest()

	if err != nil {
		log.Fatalf("Failed to get latest block: %v", err)
	}

	txIdx := exit.TxIndex.Uint64()
	lastBlockHeight := latestBlock.Header.Number
	currBlockHeight := exit.BlockNum.Uint64() + 1

	response := userclient.GetBlock(rootUrl, exit.BlockNum.Uint64())

	if txIdx >= uint64(len(response.Transactions)) {
		log.Fatalln("The following exit does not exist within this block!")
	}

	exitTx := response.Transactions[exit.TxIndex.Uint64()]

	log.Printf("Finding spends from blocks %d to %d\n", currBlockHeight, lastBlockHeight)

	// Find possible double spends in every block
	// TODO: actually in theory it should never happen in the current block.
	// Because root node will never create and submit that block.
	// Also, how do you protect against exits happenning more than once?
	for i := currBlockHeight; i <= lastBlockHeight; i++ {
		response := userclient.GetBlock(rootUrl, i)
		currTxs := response.Transactions
		rej := node.FindMatchingInputs(&exitTx, currTxs)

		if len(rej) > 0 {
			log.Printf("Found %d double spends at block %d\n", len(rej), i)
			// Always return the first one for now
			return currTxs, util.NewUint64(i), util.NewUint32(rej[0].TxIdx)
		} else {
			log.Printf("Found no double spends for block %d\n", i)
		}
	}

	return nil, nil, nil
}
