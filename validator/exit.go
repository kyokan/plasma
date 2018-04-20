package validator

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/kyokan/plasma/chain"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
)

// TODO: add an exit client to make from the command line.
func ExitStartedListener(level *db.Database, plasma *eth.PlasmaClient) {
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

				// TODO: do something with these...
				// we want to validate that this exit looks legit
				// meaning that it wasn't spent already.
				fmt.Println(event)
				exitId := event.ExitId
				exit := plasma.GetExit(exitId.Uint64())
				spend := FindSpend(plasma, exit)

				if spend != nil {
					// challenge
					plasma.ChallengeExit(
						exitId,
						// This is the tx that we want to use to prove it's spent.
						spend.block,
						spend.txs,
						spend.blocknum,
						spend.txindex,
					)
				}

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

func FindSpend(plasma *eth.PlasmaClient, exit eth.Exit) *TransactionInfo {
	// TODO: this needs to return the tx that is found to cause a double spend.
	return &TransactionInfo{}
}
