package validator

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	encoding_json "encoding/json"

	"github.com/gorilla/rpc/json"
	"github.com/kyokan/plasma/db"
	plasma_rpc "github.com/kyokan/plasma/rpc"
)

// clientResponse represents a JSON-RPC response returned to a client.
type ClientResponse struct {
	Result *encoding_json.RawMessage `json:"result"`
	Error  interface{}               `json:"error"`
	Id     uint64                    `json:"id"`
}

func RootNodeListener(rpcPort int, level *db.Database) {
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

		rootUrl := fmt.Sprintf("http://localhost:%d/rpc", rpcPort)

		fmt.Println(rootUrl)
		fmt.Println(blockNum)

		response := GetBlock(rootUrl, blockNum)

		fmt.Println(response)

		if response != nil {
			level.BlockDao.Save(response.Block)
		}

		time.Sleep(3 * time.Second)
	}
}

// TODO: make this a real client.
func GetBlock(url string, height uint64) *plasma_rpc.GetBlocksResponse {
	// TODO: this is plural but shouldn't be
	args := &plasma_rpc.GetBlocksArgs{
		Height: height,
	}
	message, err := json.EncodeClientRequest("Block.GetBlock", args)
	if err != nil {
		log.Fatalf("%s", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(message))
	if err != nil {
		log.Fatalf("%s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error in sending request to %s. %s", url, err)
	}
	defer resp.Body.Close()

	var c ClientResponse
	err = encoding_json.NewDecoder(resp.Body).Decode(&c)

	if err != nil {
		panic(err)
	}

	if c.Error == nil {
		var result plasma_rpc.GetBlocksResponse

		encoding_json.Unmarshal(*c.Result, &result)

		return &result
	}

	return nil
}
