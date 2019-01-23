package eth

import (
	"context"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/kyokan/plasma/util"
	"github.com/kyokan/plasma/plasma-mvp-rootchain/gen/contracts"
	"github.com/kyokan/plasma/chain"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
		)

const SignaturePreamble = "\x19Ethereum Signed Message:\n"

const depositFilter = "0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c"
const depositDescription = `[{"anonymous":false,"inputs":[{"indexed":false,"name":"sender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Deposit","type":"event"}]`

type StartExitOpts struct {
	Transaction      chain.Transaction
	Input 	 		 chain.Input
	Signature  		 []byte
	Proof            []byte
	ConfirmSignature []byte
	CommittedFee     *big.Int
}

type ChallengeExitOpts struct {
	StartExitOpts
	ExistingInput chain.Input
}

type Client interface {
	Balance(addr common.Address) (*big.Int, error)
	UserAddress() common.Address
	Contract() *contracts.Plasma
	SignData(data []byte) ([]byte, error)
	SubmitBlock(util.Hash, *big.Int, *big.Int, *big.Int) error
	SubmitBlocks([]util.Hash, []*big.Int, []*big.Int, *big.Int) error
	Deposit(value *big.Int, tx *chain.Transaction) error
	GetChildBlock(uint64) (merkleRoot [32]byte, NumTxns *big.Int, FeeAmount *big.Int, CreatedAt *big.Int, err error)

	StartDepositExit(nonce, committedFee *big.Int) error
	StartTransactionExit(opts *StartExitOpts) error
	StartFeeExit(*big.Int) error

	ChallengeExit(nonce *big.Int, opts *ChallengeExitOpts) error
	ChallengeFeeMismatch(nonce *big.Int, opts *ChallengeExitOpts) error

	FinalizeDepositExits() error
	FinalizeTransactionExits() error

	AddedToBalancesFilter(uint64) ([]contracts.PlasmaAddedToBalances, uint64, error)
	BlockSubmittedFilter(uint64) ([]contracts.PlasmaBlockSubmitted, uint64, error)
	DepositFilter(start uint64) ([]contracts.PlasmaDeposit, uint64, error)

	ChallengedExitFilter(uint64) ([]contracts.PlasmaChallengedExit, uint64, error)

	FinalizedExitFilter(uint64) ([]contracts.PlasmaFinalizedExit, uint64, error)

	StartedTransactionExitFilter(uint64) ([]contracts.PlasmaStartedTransactionExit, uint64, error)
	StartedDepositExitFilter(uint64) ([]contracts.PlasmaStartedDepositExit, uint64, error)
}

type DepositEvent struct {
	Sender       common.Address
	Value        *big.Int
	DepositNonce *big.Int
}

type clientState struct {
	client     *ethclient.Client
	rpc        *rpc.Client
	contract   *contracts.Plasma
	privateKey *ecdsa.PrivateKey
}

func NewClient(nodeUrl string, contractAddr string, privateKey *ecdsa.PrivateKey) (Client, error) {
	addr := common.HexToAddress(contractAddr)
	c, err := rpc.Dial(nodeUrl)
	if err != nil {
		return nil, err
	}

	client := ethclient.NewClient(c)
	contract, err := contracts.NewPlasma(addr, client)
	return &clientState{
		client:     client,
		rpc:        c,
		contract:   contract,
		privateKey: privateKey,
	}, nil
}

func (c *clientState) Balance(addr common.Address) (*big.Int, error) {
	log.Printf("Attempting to get balance for %s", util.AddressToHex(&addr))
	return c.client.BalanceAt(context.Background(), addr, nil)
}

func (c *clientState) UserAddress() common.Address {
	return crypto.PubkeyToAddress(*(c.privateKey.Public()).(*ecdsa.PublicKey))
}

func (c *clientState) Contract() *contracts.Plasma {
	return c.contract
}

func (c *clientState) SignData(data []byte) ([]byte, error) {
	hash := GethHash(data)
	signed, err := crypto.Sign(hash, c.privateKey)
	if err != nil {
		return nil, err
	}
	if len(signed) == 65 && signed[64] < 2 {
		signed[64] += 27;
	}
	return signed, nil
}

func (c *clientState) SubscribeDeposits(address common.Address, resChan chan<- DepositEvent) error {
	query := ethereum.FilterQuery{
		FromBlock: nil,
		ToBlock:   nil,
		Topics:    [][]common.Hash{{common.HexToHash(depositFilter)}},
		Addresses: []common.Address{address},
	}

	ch := make(chan types.Log)
	_, err := c.client.SubscribeFilterLogs(context.TODO(), query, ch)
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

