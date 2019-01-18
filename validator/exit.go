package validator

import (
	"log"
	"math/big"
	"time"

	"context"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/node"
	"github.com/kyokan/plasma/rpc"
	"github.com/kyokan/plasma/rpc/pb"
)

func ExitStartedListener(ctx context.Context, storage db.PlasmaStorage, ethClient eth.Client, rootClient pb.RootClient) {
	for {
		idx, err := storage.LastTransactionExitEventIdx()

		if err != nil && err.Error() != "leveldb: not found" {
			log.Fatalf("Failed to get last exit event idx: %v", err)
		}

		log.Printf("Looking for exit events at block number: %d\n", idx)

		events, lastIdx, err := ethClient.StartedTransactionExitFilter(idx)
		if err != nil {
			log.Printf("Failed to get transaction exit events: %s", err.Error())
			continue
		}

		if len(events) > 0 {
			count := uint64(0)
			/*
			TODO: It's not clear how can I get more information about an exit
			The contract uses the input to construct an exit which doesn't
			include information about the transaction being exited:
			txExits[position] = exit({
            	owner: txList[12 + 2 * txPos[2]].toAddress(),
            	amount: amount,
            	createdAt: block.timestamp,
            	state: ExitState.Pending
        	});

			for _, event := range events {
				count += 1

				exitId := event.Position

				exit, err := ethClient.Exit(exitId)
				if err != nil {
					log.Println("Caught error querying exit:", err)
					continue
				}

				txs, blockId, txId, err := FindDoubleSpend(ctx, rootClient, storage, exit)
				if err != nil {
					log.Println("caught error finding double spends:", err)
					continue
				}

				if txs != nil && txId != nil {
					opts := &eth.ChallengeExitOpts{
						ExitId:   exitId,
						Txs:      txs,
						BlockNum: *blockId,
						TxIndex:  *txId,
					}

					ethClient.ChallengeExit(opts)
					time.Sleep(3 * time.Second)

					events, _ := ethClient.ChallengeSuccessFilter(0)

					for _, event := range events {
						log.Printf("challenge success: %v", event.ExitId)
					}

					events2, _ := ethClient.ChallengeFailureFilter(0)

					for _, event := range events2 {
						log.Printf("challenge failure: %v", event.ExitId)
					}
				}


				// TODO: also if someone exits on the ethClient chain you need to
				// make sure you exit it from the root node.
				// So the root node also needs an exit listener.

				// There's a race condition where someone could try to spend
				// while an exit is happenning

				// This sort of implies that you should be validating exits
				// often, not just on notification.

				// It's not synchronized right now...
				time.Sleep(time.Second * 3)
			}
			*/

			log.Printf("Found %d exit events at from blocks %d to %d.\n", count, idx, lastIdx)

			storage.SaveTransactionExitEventIdx(lastIdx + 1)
		} else {
			log.Printf("No exit events at block %d.\n", idx)
		}

		time.Sleep(time.Second * 10)
	}
}

func FindDoubleSpend(ctx context.Context, rootClient pb.RootClient, storage db.PlasmaStorage, exit *eth.Exit) ([]chain.Transaction, *big.Int, *big.Int, error) {
	latestBlock, err := storage.LatestBlock()
	if err != nil {
		return nil, nil, nil, err
	}

	txIdx := exit.TxIndex
	lastBlockHeight := latestBlock.Header.Number
	currBlockHeight := exit.BlockNum + 1

	response, err := rootClient.GetBlock(ctx, &pb.GetBlockRequest{
		Number: exit.BlockNum,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	if txIdx >= uint32(len(response.Transactions)) {
		log.Fatalln("The following exit does not exist within this block!")
	}

	exitTx := rpc.DeserializeTx(response.Transactions[exit.TxIndex])
	log.Printf("Finding spends from blocks %d to %d\n", currBlockHeight, lastBlockHeight)

	// Find possible double spends in every block
	// TODO: actually in theory it should never happen in the current block.
	// Because root node will never create and submit that block.
	// Also, how do you protect against exits happening more than once?
	for i := currBlockHeight; i <= lastBlockHeight; i++ {
		response, err := rootClient.GetBlock(ctx, &pb.GetBlockRequest{
			Number: i,
		})
		if err != nil {
			return nil, nil, nil, err
		}
		currTxs := rpc.DeserializeTxs(response.Transactions)
		rej := node.FindMatchingInputs(exitTx, currTxs)

		if len(rej) > 0 {
			log.Printf("Found %d double spends at block %d\n", len(rej), i)
			// Always return the first one for now
			return currTxs, big.NewInt(int64(i)), rej[0].TxIdx, nil
		} else {
			log.Printf("Found no double spends for block %d\n", i)
		}
	}

	return nil, nil, nil, nil
}
