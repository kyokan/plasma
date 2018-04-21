package validator

import (
	"fmt"
	"log"
	"time"

	encoding_json "encoding/json"

	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/userclient"
)

// clientResponse represents a JSON-RPC response returned to a client.
type ClientResponse struct {
	Result *encoding_json.RawMessage `json:"result"`
	Error  interface{}               `json:"error"`
	Id     uint64                    `json:"id"`
}

func RootNodeListener(rootPort int, level *db.Database) {
	for {
		fmt.Println("Watching root node...")

		block, err := level.BlockDao.Latest()

		// what if latest is null
		if err != nil {
			panic(err)
		}

		var blockNum uint64

		if block == nil {
			blockNum = 1
		} else {
			blockNum = block.Header.Number + 1
		}

		log.Printf("Looking for block number: %d\n", blockNum)

		rootUrl := fmt.Sprintf("http://localhost:%d/rpc", rootPort)

		response := userclient.GetBlock(rootUrl, blockNum)

		if response != nil {
			log.Printf("Found block number: %d\n", blockNum)
			fmt.Println(response)
			// if not valid
			//	txs, err := level.AddressDao.UTXOs(&addr)

			// TODO: compare block with that on the plasma chain.
			level.BlockDao.Save(response.Block)
		}

		time.Sleep(3 * time.Second)
	}
}

// TODO: need a plasma GetBlock function as well.

func ValidBlock(block *chain.Block) bool {
	// TODO: compare this block with that on the plasma chain.
	// TODO: how long has it been since we created a new block in plasma.
	return false
}
