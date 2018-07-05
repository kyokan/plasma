package userclient

import (
	"bytes"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"

	encoding_json "encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/rpc/json"
	"github.com/kyokan/plasma/eth"
	plasma_rpc "github.com/kyokan/plasma/rpc"
	"github.com/kyokan/plasma/chain"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/urfave/cli.v1"
)

// clientResponse represents a JSON-RPC response returned to a client.
type ClientResponse struct {
	Result *encoding_json.RawMessage `json:"result"`
	Error  interface{}               `json:"error"`
	Id     uint64                    `json:"id"`
}

type client struct {
	RootURL string
}

type RootClient interface {
	GetBlock(height uint64) *plasma_rpc.GetBlocksResponse
	GetUTXOs(userAddress string) *plasma_rpc.GetUTXOsResponse
}

func NewRootClient(rootURL string) RootClient {
	c := client{RootURL: rootURL}
	return &c
}

func SendCLI(c *cli.Context) {
	userAddress := c.GlobalString("user-address")
	rootUrl := fmt.Sprintf("http://localhost:%d/rpc", c.Int("root-port"))
	nodeUrl := c.GlobalString("node-url")
	toAddr  := c.String("to")
	amount  := int64(c.Int("amount"))

	log.Printf("Sending amount: %d to: %s\n", amount, toAddr)

	client, err := eth.NewClient(nodeUrl)

	utxosEndpoint := "Block.GetUTXOs"
	utxosArgs := &plasma_rpc.GetUTXOsArgs{
		UserAddress: userAddress,
	}
	utxosResponse := request(rootUrl, utxosArgs, utxosEndpoint)
	if utxosResponse == nil {
		log.Printf("Failed to get UTXOs for %s", userAddress)
		return
	}
	var utxos plasma_rpc.GetUTXOsResponse
	err = encoding_json.Unmarshal(*utxosResponse, &utxos)
	if err != nil {
		log.Printf("Failed to unmarshal UTXO response: %s", err.Error())
		return
	}
	tx, err := chain.FindBestUTXOs(common.HexToAddress(userAddress), common.HexToAddress(toAddr), big.NewInt(amount), utxos.Transactions, client)
	if err != nil {
		log.Printf("Could not find a suitable input for send: %s", err.Error())
		return
	}


	sendArgs := &plasma_rpc.SendArgs{
		Transaction: *tx,
		From:   userAddress,
		To:     toAddr,
		Amount: fmt.Sprintf("%d", amount),
	}
	sendEndpoint  := "Transaction.Send"

	sendResponse := request(rootUrl, sendArgs, sendEndpoint)

	if sendResponse != nil {
		var result plasma_rpc.SendResponse

		err = encoding_json.Unmarshal(*sendResponse, &result)
		if err != nil {
			log.Printf("Failed to unmarshal response after sending transaction: %s", err.Error())
			return
		}

		log.Printf("Transaction sent with hash: %v\n", result.Transaction.Hash)
	} else {
		fmt.Println("Transction failed no repsonse given\n")
	}
}

func GetBlockCLI(c *cli.Context) {
	rootUrl := fmt.Sprintf("http://localhost:%d/rpc", c.Int("root-port"))
	height := uint64(c.Int("height"))

	log.Printf("Getting block for height: %d\n", height)

	rootClient := NewRootClient(rootUrl)
	response := rootClient.GetBlock(height)

	if response != nil {
		block := response.Block
		txs := response.Transactions

		table1 := tablewriter.NewWriter(os.Stdout)
		table1.SetHeader([]string{"Number", "BlockHash", "Merkle Root", "RLP Merkle Root", "Prev Hash"})
		table1.Append([]string{
			fmt.Sprint(block.Header.Number),
			common.ToHex(block.BlockHash),
			common.ToHex(block.Header.MerkleRoot),
			common.ToHex(block.Header.RLPMerkleRoot),
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

func (c client) GetBlock(height uint64) *plasma_rpc.GetBlocksResponse {
	args := &plasma_rpc.GetBlocksArgs{
		Height: height,
	}
	endpoint := "Block.GetBlock"

	response := request(c.RootURL, args, endpoint)

	if response != nil {
		var result plasma_rpc.GetBlocksResponse

		encoding_json.Unmarshal(*response, &result)

		return &result
	}

	return nil
}

func (c client) GetUTXOs(userAddress string) *plasma_rpc.GetUTXOsResponse {
	args := &plasma_rpc.GetUTXOsArgs{
		UserAddress: userAddress,
	}
	endpoint := "Block.GetUTXOs"

	response := request(c.RootURL, args, endpoint)

	if response != nil {
		var result plasma_rpc.GetUTXOsResponse

		encoding_json.Unmarshal(*response, &result)

		return &result
	}

	return nil
}

func request(url string, args interface{}, endpoint string) *encoding_json.RawMessage {
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
