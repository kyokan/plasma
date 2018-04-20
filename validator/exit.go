package validator

import (
	"fmt"
	"log"
	"time"

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

				fmt.Println(exit)

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
