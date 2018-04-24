package userclient

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"

	encoding_json "encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/rpc/json"
	plasma_rpc "github.com/kyokan/plasma/rpc"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

// clientResponse represents a JSON-RPC response returned to a client.
type ClientResponse struct {
	Result *encoding_json.RawMessage `json:"result"`
	Error  interface{}               `json:"error"`
	Id     uint64                    `json:"id"`
}

func SendCLI(c *cli.Context) {
	userAddress := c.GlobalString("user-address")
	rootUrl := fmt.Sprintf("http://localhost:%d/rpc", c.Int("root-port"))
	toAddr := c.String("to")
	amount := uint64(c.Int("amount"))

	fmt.Printf("Sending amount: %d to: %s\n", amount, toAddr)

	args := &plasma_rpc.SendArgs{
		From:   userAddress,
		To:     toAddr,
		Amount: fmt.Sprintf("%d", amount),
	}
	endpoint := "Transaction.Send"

	response := Request(rootUrl, args, endpoint)

	if response != nil {
		var result plasma_rpc.SendResponse

		encoding_json.Unmarshal(*response, &result)

		fmt.Printf("Transction sent with hash: %v\n", result.Transaction.Hash)
	} else {
		fmt.Println("Transction failed no repsonse given\n")
	}
}

func GetBlockCLI(c *cli.Context) {
	rootUrl := fmt.Sprintf("http://localhost:%d/rpc", c.Int("root-port"))
	height := uint64(c.Int("height"))

	fmt.Printf("Getting block for height: %d\n", height)

	response := GetBlock(rootUrl, height)

	if response != nil {
		block := response.Block
		txs := response.Transactions

		table1 := tablewriter.NewWriter(os.Stdout)
		table1.SetHeader([]string{"Number", "BlockHash", "Merkle Root", "Prev Hash"})
		table1.Append([]string{
			fmt.Sprint(block.Header.Number),
			common.ToHex(block.BlockHash),
			common.ToHex(block.Header.MerkleRoot),
			common.ToHex(block.Header.PrevHash),
		})

		table1.Render()

		table2 := tablewriter.NewWriter(os.Stdout)
		table2.SetHeader([]string{
			"Hash",
			"Block Number",
			"Tx Index",
			"Input0 BlkNum",
			"Input0 TxIdx",
			"Input0 OutIdx",
			"Input1 BlkNum",
			"Input1 TxIdx",
			"Input1 OutIdx",
			"Output0 NewOwner",
			"Output0 Amount",
			"Output1 NewOwner",
			"Output1 Amount",
		})
		for _, tx := range txs {
			table2.Append([]string{
				common.ToHex(tx.Hash()),
				fmt.Sprint(tx.BlkNum),
				fmt.Sprint(tx.TxIdx),
				fmt.Sprint(tx.Input0.BlkNum),
				fmt.Sprint(tx.Input0.TxIdx),
				fmt.Sprint(tx.Input0.OutIdx),
				fmt.Sprint(tx.Input1.BlkNum),
				fmt.Sprint(tx.Input1.TxIdx),
				fmt.Sprint(tx.Input1.OutIdx),
				tx.Output0.NewOwner.Hex(),
				fmt.Sprint(tx.Output0.Amount),
				tx.Output1.NewOwner.Hex(),
				fmt.Sprint(tx.Output1.Amount),
			})
		}

		table2.Render()
	} else {
		fmt.Println("Transaction failed no repsonse given\n")
	}
}

func GetBlock(url string, height uint64) *plasma_rpc.GetBlocksResponse {
	args := &plasma_rpc.GetBlocksArgs{
		Height: height,
	}
	endpoint := "Block.GetBlock"

	response := Request(url, args, endpoint)

	if response != nil {
		var result plasma_rpc.GetBlocksResponse

		encoding_json.Unmarshal(*response, &result)

		return &result
	}

	return nil
}

func GetUTXOs(url string, userAddress string) *plasma_rpc.GetUTXOsResponse {
	args := &plasma_rpc.GetUTXOsArgs{
		UserAddress: userAddress,
	}
	endpoint := "Block.GetUTXOs"

	response := Request(url, args, endpoint)

	if response != nil {
		var result plasma_rpc.GetUTXOsResponse

		encoding_json.Unmarshal(*response, &result)

		return &result
	}

	return nil
}

func Request(url string, args interface{}, endpoint string) *encoding_json.RawMessage {
	message, err := json.EncodeClientRequest(endpoint, args)
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
		log.Fatalf("Failed to decode to json: %v", err)
	}

	if c.Error == nil {
		return c.Result
	}

	return nil
}
