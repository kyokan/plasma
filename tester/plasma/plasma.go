package plasma

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/contracts/gen/contracts"
	"github.com/kyokan/plasma/tester"
	"github.com/kyokan/plasma/util"
)

func CreatePlasmaClient(nodeUrl string, contractAddress string) *contracts.Plasma {
	conn, err := ethclient.Dial(nodeUrl)

	if err != nil {
		log.Panic("Failed to start ETH client: ", err)
	}

	plasma, err := contracts.NewPlasma(common.HexToAddress(contractAddress), conn)

	if err != nil {
		log.Fatalf("Failed to instantiate a Token contract: %v", err)
	}

	return plasma
}

func CurrentChildBlock(
	plasma *contracts.Plasma,
	address string,
) *big.Int {
	opts := tester.CreateCallOpts(address)

	blocknum, err := plasma.CurrentChildBlock(opts)

	if err != nil {
		panic(err)
	}

	return blocknum
}

func LastExitId(
	plasma *contracts.Plasma,
	address string,
) *big.Int {
	opts := tester.CreateCallOpts(address)
	exitId, err := plasma.LastExitId(opts)

	if err != nil {
		panic(err)
	}

	return exitId
}

func Finalize(
	plasma *contracts.Plasma,
	privateKeyECDSA *ecdsa.PrivateKey,
) {
	auth := tester.CreateAuth(privateKeyECDSA)
	tx, err := plasma.Finalize(auth)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Finalize pending: 0x%x\n", tx.Hash())
}

func ChallengeExit(
	plasma *contracts.Plasma,
	privateKeyECDSA *ecdsa.PrivateKey,
	address string,
	txs []chain.Transaction,
	merkle util.MerkleTree,
	blocknum *big.Int,
	txindex *big.Int,
	exitId *big.Int,
) {
	auth := tester.CreateAuth(privateKeyECDSA)
	bytes, err := rlp.EncodeToBytes(&txs[txindex.Int64()])

	if err != nil {
		panic(err)
	}

	proof := CreateMerkleProof(merkle, txindex)

	tx, err := plasma.ChallengeExit(
		auth,
		exitId,
		blocknum,
		txindex,
		bytes,
		proof,
	)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Challenge Exit pending: 0x%x\n", tx.Hash())
}

func StartExit(
	plasma *contracts.Plasma,
	privateKeyECDSA *ecdsa.PrivateKey,
	address string,
	txs []chain.Transaction,
	merkle util.MerkleTree,
	blocknum *big.Int,
	txindex *big.Int,
) {
	auth := tester.CreateAuth(privateKeyECDSA)
	oindex := new(big.Int).SetUint64(0)
	bytes, err := rlp.EncodeToBytes(&txs[txindex.Int64()])

	if err != nil {
		panic(err)
	}

	proof := CreateMerkleProof(merkle, txindex)

	tx, err := plasma.StartExit(
		auth,
		blocknum,
		txindex,
		oindex,
		bytes,
		proof,
	)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Start Exit pending: 0x%x\n", tx.Hash())
}

func SubmitBlock(
	plasma *contracts.Plasma,
	privateKeyECDSA *ecdsa.PrivateKey,
	address string,
	txs []chain.Transaction,
	merkle util.MerkleTree,
) {
	auth := tester.CreateAuth(privateKeyECDSA)

	var root [32]byte
	copy(root[:], merkle.Root.Hash[:32])
	tx, err := plasma.SubmitBlock(auth, root)

	if err != nil {
		log.Fatalf("Failed to submit block: %v", err)
	}

	fmt.Printf("Submit block pending: 0x%x\n", tx.Hash())
}

func Deposit(
	plasma *contracts.Plasma,
	privateKeyECDSA *ecdsa.PrivateKey,
	address string,
	value int,
	t *chain.Transaction,
) {
	auth := tester.CreateAuth(privateKeyECDSA)
	auth.Value = new(big.Int).SetInt64(int64(value))

	bytes, err := rlp.EncodeToBytes(&t)

	if err != nil {
		panic(err)
	}

	tx, err := plasma.Deposit(auth, bytes)

	if err != nil {
		log.Fatalf("Failed to deposit: %v", err)
	}

	fmt.Printf("Deposit pending: 0x%x\n", tx.Hash())
}
