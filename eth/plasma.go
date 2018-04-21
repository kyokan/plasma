package eth

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kyokan/plasma/chain"
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

type Exit struct {
	Owner     common.Address
	Amount    *big.Int
	BlockNum  *big.Int
	TxIndex   *big.Int
	OIndex    *big.Int
	StartedAt *big.Int
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

	fmt.Println("**** submit block")
	fmt.Println(hex.EncodeToString(root[:]))

	if err != nil {
		log.Fatalf("Failed to submit block: %v", err)
	}

	fmt.Printf("Submit block pending: 0x%x\n", tx.Hash())
}

func (p *PlasmaClient) Deposit(
	value uint64,
	t *chain.Transaction,
) {
	var opts *bind.TransactOpts

	if p.useGeth {
		opts = p.ethClient.NewGethTransactor(common.HexToAddress(p.userAddress))
	} else {
		opts = util.CreateAuth(p.privateKey)
	}

	opts.Value = util.NewUint64(value)

	bytes, err := rlp.EncodeToBytes(&t)

	if err != nil {
		panic(err)
	}

	tx, err := p.plasma.Deposit(opts, bytes)

	if err != nil {
		log.Fatalf("Failed to deposit: %v", err)
	}

	fmt.Printf("Deposit pending: 0x%x\n", tx.Hash())
}

func (p *PlasmaClient) StartExit(
	block *chain.Block,
	txs []chain.Transaction,
	blocknum *big.Int,
	txindex *big.Int,
	oindex *big.Int,
) {
	var opts *bind.TransactOpts

	if p.useGeth {
		opts = p.ethClient.NewGethTransactor(common.HexToAddress(p.userAddress))
	} else {
		opts = util.CreateAuth(p.privateKey)
	}

	tx := txs[txindex.Int64()]

	bytes, err := rlp.EncodeToBytes(&tx)

	if err != nil {
		panic(err)
	}

	// TODO: are these hashes correct?
	merkle := CreateMerkleTree(txs)
	proof := util.CreateMerkleProof(merkle, txindex)

	fmt.Println("**** start exit")
	fmt.Printf("block: %d\n", blocknum)
	fmt.Printf("txindex: %d\n", txindex)
	fmt.Printf("oindex: %d\n", oindex)
	fmt.Println("tx: " + hex.EncodeToString(bytes))
	fmt.Println("proof: " + hex.EncodeToString(proof))
	fmt.Println("address: " + tx.Output0.NewOwner.Hex())

	res, err := p.plasma.StartExit(
		opts,
		blocknum,
		txindex,
		oindex,
		bytes,
		proof,
	)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Start Exit pending: 0x%x\n", res.Hash())
}

func (p *PlasmaClient) ChallengeExit(
	exitId *big.Int,
	block *chain.Block,
	txs []chain.Transaction,
	blocknum *big.Int,
	txindex *big.Int,
) {
	var opts *bind.TransactOpts

	if p.useGeth {
		opts = p.ethClient.NewGethTransactor(common.HexToAddress(p.userAddress))
	} else {
		opts = util.CreateAuth(p.privateKey)
	}

	tx := txs[txindex.Int64()]

	bytes, err := rlp.EncodeToBytes(&tx)

	if err != nil {
		panic(err)
	}

	merkle := CreateMerkleTree(txs)
	proof := util.CreateMerkleProof(merkle, txindex)

	res, err := p.plasma.ChallengeExit(
		opts,
		exitId,
		blocknum,
		txindex,
		bytes,
		proof,
	)

	if err != nil {
		log.Fatalf("Failed to challenge exit: %v", err)
	}

	fmt.Printf("Challenge Exit pending: 0x%x\n", res.Hash())
}

func (p *PlasmaClient) GetExit(exitId *big.Int) Exit {
	opts := util.CreateCallOpts(p.userAddress)

	owner, amount, blocknum, txindex, oindex, startedAt, err := p.plasma.GetExit(opts, exitId)

	if err != nil {
		log.Fatalf("Failed to get exit: %v", err)
	}

	return Exit{
		owner,
		amount,
		blocknum,
		txindex,
		oindex,
		startedAt,
	}
}

func (p *PlasmaClient) CurrentChildBlock() (*big.Int, error) {
	opts := util.CreateCallOpts(p.userAddress)
	return p.plasma.CurrentChildBlock(opts)
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

func (p *PlasmaClient) ExitStartedFilter(
	start uint64,
) ([]contracts.PlasmaExitStarted, uint64) {
	opts := bind.FilterOpts{
		Start:   start,
		End:     nil, // TODO: end doesn't seem to work
		Context: context.Background(),
	}

	itr, err := p.plasma.FilterExitStarted(&opts)

	if err != nil {
		panic(err)
	}

	next := true

	var events []contracts.PlasmaExitStarted

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

func (p *PlasmaClient) DebugAddressFilter(
	start uint64,
) ([]contracts.PlasmaDebugAddress, uint64) {
	opts := bind.FilterOpts{
		Start:   start, // TODO: in the future we should store the last starting point in the db.
		End:     nil,
		Context: context.Background(),
	}

	itr, err := p.plasma.FilterDebugAddress(&opts)

	if err != nil {
		panic(err)
	}

	next := true

	var events []contracts.PlasmaDebugAddress

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

func (p *PlasmaClient) DebugUintFilter(
	start uint64,
) ([]contracts.PlasmaDebugUint, uint64) {
	opts := bind.FilterOpts{
		Start:   start, // TODO: in the future we should store the last starting point in the db.
		End:     nil,
		Context: context.Background(),
	}

	itr, err := p.plasma.FilterDebugUint(&opts)

	if err != nil {
		panic(err)
	}

	next := true

	var events []contracts.PlasmaDebugUint

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

func (p *PlasmaClient) DebugBoolFilter(
	start uint64,
) ([]contracts.PlasmaDebugBool, uint64) {
	opts := bind.FilterOpts{
		Start:   0x0, // TODO: in the future we should store the last starting point in the db.
		End:     nil,
		Context: context.Background(),
	}

	itr, err := p.plasma.FilterDebugBool(&opts)

	if err != nil {
		panic(err)
	}

	next := true

	var events []contracts.PlasmaDebugBool

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

// TODO: it prevents import cycle with utils.
func CreateMerkleTree(accepted []chain.Transaction) util.MerkleTree {
	hashables := make([]util.RLPHashable, len(accepted))

	for i := range accepted {
		txPtr := &accepted[i]
		hashables[i] = util.RLPHashable(txPtr)
	}

	merkle := util.TreeFromRLPItems(hashables)
	return merkle
}
