package eth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kyokan/plasma/contracts/gen/contracts"
	"github.com/kyokan/plasma/util"
)

type PlasmaClient struct {
	plasma      *contracts.Plasma
	privateKey  *ecdsa.PrivateKey
	userAddress string
	ethClient   *Client
	useGeth     bool
}

func CreatePlasmaClient(
	nodeUrl string,
	contractAddress string,
	userAddress string,
	privateKeyECDSA *ecdsa.PrivateKey,
	useGeth bool,
) *PlasmaClient {
	conn, err := ethclient.Dial(nodeUrl)

	if err != nil {
		log.Panic("Failed to start ETH client: ", err)
	}

	plasma, err := contracts.NewPlasma(common.HexToAddress(contractAddress), conn)

	if err != nil {
		log.Fatalf("Failed to instantiate a Token contract: %v", err)
	}

	if privateKeyECDSA == nil {
		panic("Private key ecdsa not found")
	}

	ethClient, err := NewClient(nodeUrl)

	if err != nil {
		log.Panic("Failed to create a new eth client", err)
	}

	return &PlasmaClient{
		plasma,
		privateKeyECDSA,
		userAddress,
		ethClient,
		useGeth,
	}
}

func (p *PlasmaClient) SubmitBlock(
	merkle util.MerkleTree,
) {
	var opts *bind.TransactOpts

	if p.useGeth {
		opts = p.ethClient.NewGethTransactor(common.HexToAddress(p.userAddress))
	} else {
		opts = util.CreateAuth(p.privateKey)
	}

	var root [32]byte
	copy(root[:], merkle.Root.Hash[:32])
	tx, err := p.plasma.SubmitBlock(opts, root)

	if err != nil {
		log.Fatalf("Failed to submit block: %v", err)
	}

	fmt.Printf("Submit block pending: 0x%x\n", tx.Hash())
}

func (p *PlasmaClient) DepositFilter(
	start uint64,
) ([]contracts.PlasmaDeposit, uint64) {
	opts := bind.FilterOpts{
		Start:   start,
		End:     nil, // TODO: end doesn't seem to work
		Context: context.Background(),
	}

	itr, err := p.plasma.FilterDeposit(&opts)

	if err != nil {
		panic(err)
	}

	next := true

	var events []contracts.PlasmaDeposit

	var lastBlockNumber uint64

	for next {
		if itr.Event != nil {
			lastBlockNumber = itr.Event.Raw.BlockNumber
			events = append(events, *itr.Event)
		}
		next = itr.Next()
	}

	return events, lastBlockNumber
}
