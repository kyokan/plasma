package node

import (
	"log"
	"time"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
)

func StartDepositListener(storage db.PlasmaStorage, sink *TransactionSink, plasma *eth.PlasmaClient) {
	ch := make(chan eth.DepositEvent)
	sink.AcceptDepositEvents(ch)

	for {
		idx, err := storage.LastDepositEventIdx()

		if err != nil && err.Error() != "leveldb: not found" {
			log.Fatalf("Failed to get last deposit event idx: %v", err)
		}

		log.Printf("Looking for deposit events at block number: %d\n", idx)

		events, lastIdx := plasma.DepositFilter(idx)

		if len(events) > 0 {
			count := uint64(0)

			for _, event := range events {
				ch <- eth.DepositEvent{
					Sender: event.Sender,
					Value:  event.Value,
				}

				count++

				// TODO: AcceptDepositEvents is not synchronized so sleeps are required.
				time.Sleep(time.Second * 3)
			}

			log.Printf("Found %d deposit events at from blocks %d to %d.\n", count, idx, lastIdx)

			storage.SaveDepositEventIdx(lastIdx + 1)
		} else {
			log.Printf("No deposit events at block %d.\n", idx)
		}

		// Every 10 seconds look for deposits
		time.Sleep(time.Second * 10)
	}
}
