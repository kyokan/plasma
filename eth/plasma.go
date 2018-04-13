package eth

import (
	"crypto/ecdsa"
	"fmt"
	"log"

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
}

func CreatePlasmaClient(
	nodeUrl string,
	contractAddress string,
	userAddress string,
	privateKeyECDSA *ecdsa.PrivateKey,
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

	// And another eth client?

	ethClient, err := NewClient(nodeUrl)

	if err != nil {
		log.Panic("Failed to create a new eth client", err)
	}

	return &PlasmaClient{
		plasma,
		privateKeyECDSA,
		userAddress,
		ethClient,
	}
}

func (p *PlasmaClient) SubmitBlock(
	merkle util.MerkleTree,
) {
	// TODO: if the geth node is unlocked for the user
	// can we send tx without the private key?
	// auth := util.CreateAuth(p.privateKey)
	opts := p.ethClient.NewGethTransactor(common.HexToAddress(p.userAddress))

	var root [32]byte
	copy(root[:], merkle.Root.Hash[:32])
	tx, err := p.plasma.SubmitBlock(opts, root)

	if err != nil {
		log.Fatalf("Failed to submit block: %v", err)
	}

	fmt.Printf("Submit block pending: 0x%x\n", tx.Hash())
}
