package validator

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/kyokan/plasma/merkle"
	"log"
	"math/big"
	"time"

	"encoding/json"

	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/rpc"
	"github.com/kyokan/plasma/rpc/pb"
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
			Number: blockNum,
		})
		if err != nil {
			log.Println("caught error getting block", err)
			continue
		}

		if response != nil {
			log.Printf("Found block number: %d\n", blockNum)
			plasmaBlock := rpc.DeserializeBlock(response.Block)

			root, _, _, created, err := ethClient.GetChildBlock(blockNum)
			if err != nil {
				log.Println("caught error getting block", err)
			}

			if IsValidBlock(plasmaBlock, root, created) {
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
	res, err := rootClient.GetOutputs(ctx, &pb.GetOutputsRequest{
		Address: userAddress.Bytes(),
	})
	if err != nil {
		return err
	}

	confirmedTransactions := rpc.DeserializeConfirmedTxs(res.ConfirmedTransactions)

	type UTXO struct {
		chain.Transaction
		outputIdx *big.Int
	}

	utxosByBlock := make(map[*big.Int][]UTXO)

	for _, confirmed := range confirmedTransactions {
		utxos := utxosByBlock[confirmed.Transaction.BlkNum]

		if utxos == nil {
			utxos = []UTXO{}
		}

		// Collect a list of outputs because technically both can belong to the user.
		var outputIdxs []*big.Int

		if confirmed.Transaction.Output0.Owner == userAddress {
			outputIdxs = append(outputIdxs, chain.Zero())
		} else if confirmed.Transaction.Output1.Owner == userAddress {
			outputIdxs = append(outputIdxs, chain.One())
		} else {
			log.Fatalf("Transaction must have at least one output that belongs to address: %s\n", userAddress)
		}

		for _, outputIdx := range outputIdxs {
			utxo := UTXO{outputIdx: outputIdx}
			utxo.Transaction = confirmed.Transaction
			utxos = append(utxos, utxo)
		}

		utxosByBlock[confirmed.Transaction.BlkNum] = utxos
	}

	// TODO: This is highly inefficient, needs to be fixed
	for blkNum, utxos := range utxosByBlock {
		res2, err := rootClient.GetBlock(ctx, &pb.GetBlockRequest{
			Number: blkNum.Uint64(),
		})
		if err != nil {
			return err
		}
		transactions := rpc.DeserializeConfirmedTxs(res2.ConfirmedTransactions)
		hashables := make([]merkle.DualHashable, len(transactions))
		for i := 0; i < len(transactions); i++ {
			hashables[i] = &transactions[i]
		}
		for _, utxo := range utxos {
			log.Printf("Exiting block: %s, tx: %s, output: %s\n", utxo.BlkNum.String(), utxo.TxIdx.String(), utxo.outputIdx.String())
			proof, _ := merkle.GetProof(hashables, merkle.DefaultDepth, int32(utxo.TxIdx.Int64()))
			var address common.Address
			opts := &eth.StartExitOpts{
				Transaction: utxo.Transaction,
				Input: *chain.NewInput(utxo.BlkNum, utxo.TxIdx, utxo.outputIdx, chain.Zero(), address),
				Signature: []byte{}, // TODO: Fix this
				ConfirmSignature: []byte{}, // TODO: Fix this
				Proof: proof,
			}
			ethClient.StartTransactionExit(opts)

			time.Sleep(3 * time.Second)
		}
	}

	return nil
}

func IsValidBlock(block *chain.Block, root [32]byte, created *big.Int) bool {
	fmt.Println(block.Header.Number)
	fmt.Println(hex.EncodeToString(block.Header.RLPMerkleRoot))
	fmt.Println(hex.EncodeToString(root[:]))
	return bytes.Equal(block.Header.RLPMerkleRoot, root[:])
}
