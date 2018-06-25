package eth

import (
	"crypto/ecdsa"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/rlp"
	"gopkg.in/urfave/cli.v1"

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
	ethClient   Client
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

type Block struct {
	Root      []byte
	StartedAt *big.Int
}

func CreatePlasmaClientCLI(c *cli.Context) *PlasmaClient {
	contractAddress := c.GlobalString("contract-addr")
	nodeURL := c.GlobalString("node-url")
	keystoreDir := c.GlobalString("keystore-dir")
	keystoreFile := c.GlobalString("keystore-file")
	userAddress := c.GlobalString("user-address")
	privateKey := c.GlobalString("private-key")
	signPassphrase := c.GlobalString("sign-passphrase")
	useGeth := c.GlobalBool("use-geth")

	privateKeyECDSA := util.CreatePrivateKeyECDSA(
		userAddress,
		privateKey,
		keystoreDir,
		keystoreFile,
		signPassphrase,
	)

	return CreatePlasmaClient(
		nodeURL,
		contractAddress,
		userAddress,
		privateKeyECDSA,
		useGeth,
	)
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
		log.Fatalf("Failed to start ETH client: %v", err)
	}

	plasma, err := contracts.NewPlasma(common.HexToAddress(contractAddress), conn)

	if err != nil {
		log.Fatalf("Failed to instantiate a Token contract: %v", err)
	}

	if privateKeyECDSA == nil {
		log.Fatalln("Private key ecdsa not found")
	}

	// TODO: this is a duplicate eth client, might be able to merge them.
	ethClient, err := NewClient(nodeUrl)

	if err != nil {
		log.Fatalf("Failed to create a new eth client: %v", err)
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

	log.Printf("Submit block pending: 0x%x\n", tx.Hash())
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
		log.Fatalf("Failed to encode tx to rlp bytes: %v", err)
	}

	tx, err := p.plasma.Deposit(opts, bytes)

	if err != nil {
		log.Fatalf("Failed to deposit: %v", err)
	}

	log.Printf("Deposit pending: 0x%x\n", tx.Hash())
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
		log.Fatalf("Failed to encode tx to rlp bytes: %v", err)
	}

	merkle := CreateMerkleTree(txs)
	proof := util.CreateMerkleProof(merkle, txindex)

	res, err := p.plasma.StartExit(
		opts,
		blocknum,
		txindex,
		oindex,
		bytes,
		proof,
	)

	if err != nil {
		log.Fatalf("Failed to start exit: %v", err)
	}

	log.Printf("Start Exit pending: 0x%x\n", res.Hash())
}

func (p *PlasmaClient) ChallengeExit(
	exitId *big.Int,
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
		log.Fatalf("Failed to encode tx to rlp bytes: %v", err)
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

	log.Printf("Challenge Exit pending: 0x%x\n", res.Hash())
}

func (p *PlasmaClient) Finalize() {
	var opts *bind.TransactOpts

	if p.useGeth {
		opts = p.ethClient.NewGethTransactor(common.HexToAddress(p.userAddress))
	} else {
		opts = util.CreateAuth(p.privateKey)
	}

	res, err := p.plasma.Finalize(opts)

	if err != nil {
		log.Fatalf("Failed to finalize exits: %v", err)
	}

	log.Printf("Finalize pending: 0x%x\n", res.Hash())
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

func (p *PlasmaClient) GetBlock(blocknum *big.Int) Block {
	opts := util.CreateCallOpts(p.userAddress)

	log.Printf("GetBlock for address 0x%x\n", opts.From)
	root, startedAt, err := p.plasma.GetBlock(opts, blocknum)

	if err != nil {
		log.Fatalf("Failed to get block: %v", err)
	}

	return Block{
		root[:],
		startedAt,
	}
}

func (p *PlasmaClient) CurrentChildBlock() (*big.Int, error) {
	opts := util.CreateCallOpts(p.userAddress)
	return p.plasma.CurrentChildBlock(opts)
}

// Note this prevents import cycle with utils.
func CreateMerkleTree(accepted []chain.Transaction) util.MerkleTree {
	hashables := make([]util.RLPHashable, len(accepted))

	for i := range accepted {
		txPtr := &accepted[i]
		hashables[i] = util.RLPHashable(txPtr)
	}

	merkle := util.TreeFromRLPItems(hashables)
	return merkle
}
