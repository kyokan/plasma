package eth

import (
	"context"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/kyokan/plasma/util"
)

const depositFilter = "0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c"
const depositDescription = `[{"anonymous":false,"inputs":[{"indexed":false,"name":"sender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Deposit","type":"event"}]`

type DepositEvent struct {
	Sender common.Address
	Value  *big.Int
}

type Client struct {
	typedClient *ethclient.Client
	rpcClient   *rpc.Client
}

func NewClient(url string) (*Client, error) {
	c, err := rpc.Dial(url)

	if err != nil {
		return nil, err
	}

	return &Client{typedClient: ethclient.NewClient(c), rpcClient: c}, nil
}

func (c *Client) GetBalance(addr common.Address) (*big.Int, error) {
	log.Printf("Attempting to get balance for %s", util.AddressToHex(&addr))

	return c.typedClient.BalanceAt(context.Background(), addr, nil)
}

func (c *Client) SignData(addr *common.Address, data []byte) ([]byte, error) {
	log.Printf("Attempting to sign data on behalf of %s", util.AddressToHex(addr))
	var res []byte
	err := c.rpcClient.Call(&res, "eth_sign", util.AddressToHex(addr), common.ToHex(data))
	log.Printf("Received signature on behalf of %s.", util.AddressToHex(addr))

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Can be used by plasma client to send a sign transaction request to a remote geth node.
// TODO: needs to be tested.
func (c *Client) NewGethTransactor(keyAddr common.Address) *bind.TransactOpts {
	return &bind.TransactOpts{
		From: keyAddr,
		Signer: func(signer types.Signer, address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			data := signer.Hash(tx).Bytes()
			signature, err := c.SignData(&address, data)
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(signer, signature)
		},
	}
}

func (c *Client) SubscribeDeposits(address common.Address, resChan chan<- DepositEvent) error {
	query := ethereum.FilterQuery{
		FromBlock: nil,
		ToBlock:   nil,
		Topics:    [][]common.Hash{{common.HexToHash(depositFilter)}},
		Addresses: []common.Address{address},
	}

	ch := make(chan types.Log)
	_, err := c.typedClient.SubscribeFilterLogs(context.TODO(), query, ch)

	if err != nil {
		return err
	}

	log.Printf("Watching for deposits on address %s.", util.AddressToHex(&address))

	depositAbi, err := abi.JSON(strings.NewReader(depositDescription))

	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event := <-ch:
				parseDepositEvent(&depositAbi, resChan, &event)
			}
		}
	}()

	return nil
}

func parseDepositEvent(depositAbi *abi.ABI, resChan chan<- DepositEvent, raw *types.Log) {
	event := DepositEvent{}
	err := depositAbi.Unpack(&event, "Deposit", raw.Data)

	if err != nil {
		log.Print("Failed to unpack deposit: ", err)
		return
	}

	log.Printf("Received %s wei deposit from %s.", event.Value.String(), util.AddressToHex(&event.Sender))
	resChan <- event
}
