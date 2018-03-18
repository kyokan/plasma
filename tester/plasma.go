package tester

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/contracts/gen/contracts"
	"github.com/kyokan/plasma/util"
)

func StartExit(
	plasma *contracts.Plasma,
	privateKeyECDSA *ecdsa.PrivateKey,
) {
	auth := createAuth(privateKeyECDSA)

	blocknum := new(big.Int).SetUint64(1)
	txindex := new(big.Int).SetUint64(1)
	oindex := new(big.Int).SetUint64(1)

	// TODO: use cbor instead
	t := &chain.Transaction{
		Input0: chain.ZeroInput(),
		Input1: chain.ZeroInput(),
		Sig0:   []byte{},
		Sig1:   []byte{},
		Output0: &chain.Output{
			NewOwner: common.HexToAddress("0x627306090abaB3A6e1400e9345bC60c78a8BEf57"),
			Amount:   new(big.Int).SetUint64(100),
		},
		Output1: chain.ZeroOutput(),
		Fee:     new(big.Int),
		BlkNum:  uint64(1),
		TxIdx:   0,
	}

	bytes, err := t.ToCbor()

	if err != nil {
		panic(err)
	}

	proof := []byte{}

	tx, err := plasma.StartExit(auth, blocknum, txindex, oindex, bytes, proof)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Start Exit pending: 0x%x\n", tx.Hash())
}

func SubmitBlock(
	plasma *contracts.Plasma,
	privateKeyECDSA *ecdsa.PrivateKey,
) {
	auth := createAuth(privateKeyECDSA)

	root := createMerkleRoot()

	tx, err := plasma.SubmitBlock(auth, root)

	if err != nil {
		log.Fatalf("Failed to submit block: %v", err)
	}

	fmt.Printf("Submit block pending: 0x%x\n", tx.Hash())
}

func Deposit(
	plasma *contracts.Plasma,
	privateKeyECDSA *ecdsa.PrivateKey,
	value uint64,
) {
	auth := createAuth(privateKeyECDSA)
	auth.Value = new(big.Int).SetUint64(value)

	// TODO: use cbor instead
	t := &chain.Transaction{
		Input0: chain.ZeroInput(),
		Input1: chain.ZeroInput(),
		Sig0:   []byte{},
		Sig1:   []byte{},
		Output0: &chain.Output{
			NewOwner: common.HexToAddress("0x627306090abaB3A6e1400e9345bC60c78a8BEf57"),
			Amount:   new(big.Int).SetUint64(100),
		},
		Output1: chain.ZeroOutput(),
		Fee:     new(big.Int),
		BlkNum:  uint64(1),
		TxIdx:   0,
	}

	bytes, err := t.ToCbor()

	if err != nil {
		panic(err)
	}

	tx, err := plasma.Deposit(auth, bytes)

	if err != nil {
		log.Fatalf("Failed to deposit: %v", err)
	}

	fmt.Printf("Deposit pending: 0x%x\n", tx.Hash())
}

func createAuth(privateKeyECDSA *ecdsa.PrivateKey) *bind.TransactOpts {
	auth := bind.NewKeyedTransactor(privateKeyECDSA)
	auth.GasPrice = new(big.Int).SetUint64(1)
	auth.GasLimit = uint64(4712388)
	return auth
}

func createMerkleRoot() [32]byte {
	blkNum := uint64(1)

	accepted := []chain.Transaction{
		chain.Transaction{
			Input0: chain.ZeroInput(),
			Input1: chain.ZeroInput(),
			Sig0:   []byte{},
			Sig1:   []byte{},
			Output0: &chain.Output{
				NewOwner: common.HexToAddress("0x627306090abaB3A6e1400e9345bC60c78a8BEf57"),
				Amount:   new(big.Int).SetUint64(100),
			},
			Output1: chain.ZeroOutput(),
			Fee:     new(big.Int),
			BlkNum:  blkNum,
			TxIdx:   0,
		},
	}

	hashables := make([]util.Hashable, len(accepted))

	for i := range accepted {
		txPtr := &accepted[i]
		txPtr.BlkNum = blkNum
		txPtr.TxIdx = uint32(i)
		hashables[i] = util.Hashable(txPtr)
	}

	merkle := util.TreeFromItems(hashables)

	var res [32]byte
	hash := merkle.Root.Hash

	for i := 0; i < Min(len(res), len(hash)); i++ {
		res[i] = hash[i]
	}

	return res
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
