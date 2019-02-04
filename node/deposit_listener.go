package node

import (
	"github.com/kyokan/plasma/eth/contracts"
	"log"
	"time"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/chain"
	"math/big"
)

func StartDepositListener(storage db.PlasmaStorage, plasma eth.Client, mPool *Mempool) {
	var events []contracts.PlasmaDeposit
	var event contracts.PlasmaDeposit
	var lastPolledBlock uint64
	var err error

	lastPolledBlock, err = storage.LastDepositEventIdx()
	if err != nil {
		log.Panicln("failed to get last deposit event idx", err)
	}

	lastPolledBlock++

	for {
		head, err := plasma.EthereumBlockHeight()
		if err != nil {
			log.Println("failed to fetch latest ETH block")
			time.Sleep(time.Second)
			continue
		}

		if lastPolledBlock >= head {
			log.Println("no new blocks")
			time.Sleep(time.Second)
			continue
		}

		log.Printf("Looking for deposit for blocks: %d, %d\n", lastPolledBlock, head)
		events, _, err = plasma.DepositFilter(lastPolledBlock, head)
		if err != nil {
			log.Println("caught error filtering deposits", err)
			time.Sleep(time.Second)
			continue
		}

		if len(events) > 0 {
			log.Printf("Found %d deposit events from blocks %d to %d.\n", len(events), lastPolledBlock, head)

			count := uint64(0)

			for _, event = range events {
				tx := chain.Transaction{
					Input0: chain.ZeroInput(),
					Input1: chain.ZeroInput(),
					Output0: &chain.Output{
						Owner:        event.Depositor,
						Denom:        event.Amount,
						DepositNonce: event.DepositNonce,
					},
					Output1: chain.ZeroOutput(),
					Fee:     big.NewInt(0),
				}
				confirmed := chain.ConfirmedTransaction{Transaction: tx,}
				if err := mPool.Append(confirmed); err != nil {
				    log.Println("error while adding to mempool", err)
				}

				count++
			}
		} else {
			log.Printf("No deposit events at block %d.\n", lastPolledBlock)
		}

		if err = storage.SaveDepositEventIdx(head); err != nil {
			log.Printf("failed to save deposit idx: %s", err)
		}

		lastPolledBlock = head + 1
		time.Sleep(time.Second)
	}
}
