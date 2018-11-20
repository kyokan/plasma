package node

import (
	"log"
	"time"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
)

func StartDepositListener(storage db.PlasmaStorage, sink *TransactionSink, plasma eth.Client) {
	ch := make(chan eth.DepositEvent)
	sink.AcceptDepositEvents(ch)

	for {
		lastPolledIdx, err := storage.LastDepositEventIdx()

		if err != nil && err.Error() != "leveldb: not found" {
			log.Fatalf("Failed to get last deposit event idx: %v", err)
		}

		if lastPolledIdx == 0 {
			lastPolledIdx = 1
		}

		log.Printf("Looking for deposit events at block number: %d\n", lastPolledIdx)

		events, topBlockIdx, err := plasma.DepositFilter(lastPolledIdx)
		if err == nil {
			if len(events) > 0 {
				count := uint64(0)

				for _, event := range events {
					// TODO: Add deposit nonce to DepositEvent
					ch <- eth.DepositEvent{
						Sender: event.Depositor,
						Value:  event.Amount,
					}

					count++

					// TODO: AcceptDepositEvents is not synchronized so sleeps are required.
					time.Sleep(time.Second * 3)
				}

				log.Printf("Found %d deposit events at from blocks %d to %d.\n", count, lastPolledIdx, topBlockIdx)
			} else {
				log.Printf("No deposit events at block %d.\n", lastPolledIdx)
			}
			if err = storage.SaveDepositEventIdx(topBlockIdx + 1); err != nil {
				log.Printf("failed to save deposit idx: %s", err)
			}
		} else {
			log.Println("caught error filtering deposits:", err)
		}

		// Every 10 seconds look for deposits
		time.Sleep(time.Second * 10)
	}
}
