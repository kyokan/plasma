package node

import (
	"context"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"log"
	"time"
	"github.com/kyokan/plasma/util"
)

func StartExitListener(storage db.PlasmaStorage, plasma eth.Client, ctx context.Context) {
	timer := time.NewTimer(time.Second * 10)
	for {
		select {
		case <- ctx.Done():
			timer.Stop()
			return;
		case <- timer.C:
			go queryExitEvents(storage, plasma)
			timer.Reset(time.Second * 10)
		}
	}
}

func queryExitEvents(storage db.PlasmaStorage, plasma eth.Client) {
	txExitIdx, err := storage.LastTransactionExitEventIdx()

	if err != nil && err.Error() != "leveldb: not found" {
		log.Fatalf("Failed to get last transaction exit event idx: %v", err)
	}

	if txExitIdx == 0 {
		txExitIdx = 1
	}

	depExitIdx, err := storage.LastDepositExitEventIdx()
	if err != nil && err.Error() != "leveldb: not found" {
		log.Fatalf("Failed to get last deposit exit event idx: %v", err)
	}

	if depExitIdx == 0 {
		depExitIdx = 1
	}

	log.Printf("Looking for transaction exit events at block number: %d\n", txExitIdx)
	log.Printf("Looking for deposit exit events at block number: %d\n", depExitIdx)

	txExitEvents, txExitBlockIdx, err := plasma.StartedTransactionExitFilter(txExitIdx)
	inputs := make([]chain.Input, 0, 10)
	if err != nil {
		log.Println("caught error when filtering transaction exit events: ", err)
	} else {
		if len(txExitEvents) > 0 {
			count := uint64(0)
			for _, txExitEvent := range txExitEvents {
				input := chain.ZeroInput()
				input.BlkNum = util.Big2Uint64(txExitEvent.Position[0])
				input.TxIdx  = util.Big2Uint32(txExitEvent.Position[1])
				input.OutIdx = util.Big2Uint8(txExitEvent.Position[2])
				input.Owner  = txExitEvent.Owner
				count++
				inputs = append(inputs, *input)
				log.Printf("Found %d transaction exit events at from blocks %d to %d.\n", count, txExitIdx, txExitBlockIdx)
			}
		} else {
			log.Printf("No transaction exit events at block %d.\n", txExitIdx)
		}
		if err = storage.SaveTransactionExitEventIdx(txExitBlockIdx + 1); err != nil {
			log.Printf("failed to save transaction event idx: %s", err)
		}
	}

	depExitEvents, depExitBlockIdx, err := plasma.StartedDepositExitFilter(depExitIdx)
	if err != nil {
		log.Println("caught error when filtering deposit exit events: ", err)
	} else {
		if len(depExitEvents) > 0 {
			count := uint64(0)
			for _, depExitEvent := range depExitEvents {
				input := chain.ZeroInput()
				input.DepositNonce = depExitEvent.Nonce
				input.Owner = depExitEvent.Owner
				count++
				inputs = append(inputs, *input)
				log.Printf("Found %d deposit exit events at from blocks %d to %d.\n", count, depExitIdx, depExitBlockIdx)
			}
		} else {
			log.Printf("No deposit exit events at block %d.\n", depExitIdx)
		}
		if err = storage.SaveDepositExitEventIdx(depExitBlockIdx+ 1); err != nil {
			log.Printf("failed to save deposit event idx: %s", err)
		}
	}
	if len(inputs) > 0 {
		err = storage.MarkExitsAsSpent(inputs)
		if err != nil {
			log.Printf("failed to process exits %s", err)
		}
	}
}