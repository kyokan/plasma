package validator

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"encoding/json"

	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/util"
	"github.com/kyokan/plasma/rpc/pb"
	"github.com/kyokan/plasma/rpc"
	"context"
	"github.com/ethereum/go-ethereum/common"
)

// clientResponse represents a JSON-RPC response returned to a client.
type ClientResponse struct {
	Result *json.RawMessage `json:"result"`
	Error  interface{}      `json:"error"`
	Id     uint64           `json:"id"`
}

func RootNodeListener(ctx context.Context, storage db.PlasmaStorage, ethClient eth.Client, rootClient pb.RootClient, userAddress common.Address) {
	for {
		log.Println("Watching root node...")

		block, err := storage.LatestBlock()

		if err != nil {
			log.Fatalf("Failed to get latest block: %v", err)
		}

		var blockNum uint64

		if block == nil { // first block in the plasma chain is genesis
			blockNum = 1
		} else {
			blockNum = block.Header.Number + 1
		}

		log.Printf("Looking for block number: %d\n", blockNum)

		response, err := rootClient.GetBlock(ctx, &pb.GetBlockRequest{
			Number: rpc.SerializeBig(util.NewUint64(blockNum)),
		})
		if err != nil {
			log.Println("caught error getting block", err)
			continue
		}

		if response != nil {
			log.Printf("Found block number: %d\n", blockNum)
			plasmaBlock := rpc.DeserializeBlock(response.Block)

			// Block number for the contract is off by one
			contractBlock, err := ethClient.Block(util.NewUint64(blockNum))
			if err != nil {
				log.Println("caught error getting block", err)
			}

			if IsValidBlock(plasmaBlock, *contractBlock) {
				log.Println("Block is valid, saving locally.")
				storage.SaveBlock(plasmaBlock)
			} else {
				_, err := storage.GetInvalidBlock(plasmaBlock.BlockHash)

				if err != nil && err.Error() == "leveldb: not found" {
					log.Println("Block is not valid, starting exit of utxos.")
					// Exit all utxos, because we suspect root node is dishonest
					err = ExitUTXOs(ctx, ethClient, rootClient, userAddress)
					if err != nil {
						log.Println("caught error getting UTXOs to exit", err)
						continue
					}

					log.Println("Saving invalid block...")

					storage.SaveInvalidBlock(plasmaBlock)
				} else {
					log.Println("We already invalidated this block")
				}
			}
		}

		// Need to wait longer, because we need to wait for block to be submitted.
		time.Sleep(10 * time.Second)
	}
}

func ExitUTXOs(ctx context.Context, ethClient eth.Client, rootClient pb.RootClient, userAddress common.Address) error {
	res, err := rootClient.GetUTXOs(ctx, &pb.GetUTXOsRequest{
		Address: userAddress.Bytes(),
	})
	if err != nil {
		return err
	}

	txs := rpc.DeserializeTxs(res.Transactions)

	type UTXO struct {
		BlkNum    uint64
		TxIdx     uint32
		OutputIdx int
	}

	utxosByBlock := make(map[uint64][]UTXO)

	for _, tx := range txs {
		utxos := utxosByBlock[tx.BlkNum]

		if utxos == nil {
			utxos = []UTXO{}
		}

		// Collect a list of outputs because technically both can belong to the user.
		var outputIdxs []int

		if tx.Output0.NewOwner == userAddress {
			outputIdxs = append(outputIdxs, 0)
		} else if tx.Output1.NewOwner == userAddress {
			outputIdxs = append(outputIdxs, 1)
		} else {
			log.Fatalf("Transaction must have at least one output that belongs to address: %s\n", userAddress)
		}

		for _, outputIdx := range outputIdxs {
			utxos = append(
				utxos,
				UTXO{
					tx.BlkNum,
					tx.TxIdx,
					outputIdx,
				},
			)
		}

		utxosByBlock[tx.BlkNum] = utxos
	}

	for blkNum, utxos := range utxosByBlock {
		res2, err := rootClient.GetBlock(ctx, &pb.GetBlockRequest{
			Number: rpc.SerializeBig(util.NewUint64(blkNum)),
		})
		if err != nil {
			return err
		}

		for _, utxo := range utxos {
			log.Printf("Exiting block: %d, tx: %d, output: %d\n", utxo.BlkNum, utxo.TxIdx, utxo.OutputIdx)

			ethClient.StartExit(&eth.StartExitOpts{
				Block:    rpc.DeserializeBlock(res2.Block),
				Txs:      rpc.DeserializeTxs(res2.Transactions),
				BlockNum: util.NewUint64(utxo.BlkNum),
				TxIndex:  uint(utxo.TxIdx),
				OutIndex: uint(utxo.OutputIdx),
			})

			time.Sleep(3 * time.Second)
		}
	}

	return nil
}

func IsValidBlock(block *chain.Block, plasmaBlock eth.Block) bool {
	fmt.Println(block.Header.Number)
	fmt.Println(hex.EncodeToString(block.Header.RLPMerkleRoot))
	fmt.Println(hex.EncodeToString(plasmaBlock.Root))
	return bytes.Equal(block.Header.RLPMerkleRoot, plasmaBlock.Root)
}
