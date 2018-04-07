package plasma

import (
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/contracts/gen/contracts"
	"github.com/kyokan/plasma/util"
	"github.com/urfave/cli"
)

func IntegrationTest(c *cli.Context) {
	contractAddress := c.GlobalString("contract-addr")
	nodeURL := c.GlobalString("node-url")
	keystoreDir := c.GlobalString("keystore-dir")
	keystoreFile := c.GlobalString("keystore-file")
	userAddress := c.GlobalString("user-address")
	privateKey := c.GlobalString("private-key")
	signPassphrase := c.GlobalString("sign-passphrase")

	var privateKeyECDSA *ecdsa.PrivateKey

	if exists(userAddress) && exists(privateKey) {
		privateKeyECDSA = util.ToPrivateKeyECDSA(privateKey)
	} else if exists(keystoreDir) &&
		exists(keystoreFile) &&
		exists(userAddress) {
		keyWrapper := util.GetFromKeyStore(userAddress, keystoreDir, keystoreFile, signPassphrase)
		privateKeyECDSA = keyWrapper.PrivateKey
	}

	if privateKeyECDSA == nil {
		panic("Private key ecdsa not found")
	}

	plasma := CreatePlasmaClient(nodeURL, contractAddress)

	exitAndChallengeSameBlock(plasma, privateKeyECDSA, userAddress)
	exitAndChallengeDeposit(plasma, privateKeyECDSA, userAddress)
	finalize(plasma, privateKeyECDSA, userAddress)
	logs(plasma)
}

func exitAndChallengeSameBlock(
	plasma *contracts.Plasma,
	privateKeyECDSA *ecdsa.PrivateKey,
	userAddress string,
) {
	// Current block is the next block. Should we change that?
	blocknum := CurrentChildBlock(plasma, userAddress)
	txs := createSubmitBlockTxs(blocknum, userAddress)
	merkle := CreateMerkleTree(txs)

	SubmitBlock(plasma, privateKeyECDSA, userAddress, txs, merkle)
	time.Sleep(5 * time.Second)

	blocknum = CurrentChildBlock(plasma, userAddress)
	StartExit(
		plasma,
		privateKeyECDSA,
		userAddress,
		txs,
		merkle,
		util.Sub(blocknum, 1),
		new(big.Int).SetInt64(1),
	)
	time.Sleep(3 * time.Second)

	exitId := LastExitId(plasma, userAddress)
	ChallengeExit(
		plasma,
		privateKeyECDSA,
		userAddress,
		txs,
		merkle,
		util.Sub(blocknum, 1),
		new(big.Int).SetInt64(2),
		exitId,
	)
	time.Sleep(3 * time.Second)
}

func exitAndChallengeDeposit(
	plasma *contracts.Plasma,
	privateKeyECDSA *ecdsa.PrivateKey,
	userAddress string,
) {
	depositValue := 1000000000

	t := createDepositTx(userAddress, depositValue)
	depositMerkle := CreateMerkleTree([]chain.Transaction{t})
	Deposit(plasma, privateKeyECDSA, userAddress, depositValue, &t)
	time.Sleep(5 * time.Second)

	blocknum := CurrentChildBlock(plasma, userAddress)

	// Transactions for next block
	txs := createSubmitBlockTxs(blocknum, userAddress)
	merkle := CreateMerkleTree(txs)

	SubmitBlock(plasma, privateKeyECDSA, userAddress, txs, merkle)
	time.Sleep(5 * time.Second)

	blocknum = CurrentChildBlock(plasma, userAddress)
	StartExit(
		plasma,
		privateKeyECDSA,
		userAddress,
		[]chain.Transaction{t},
		// But this shouldn't work because the merkle for the deposit is diff.
		depositMerkle,
		// Exit deposit which is the previous block
		util.Sub(blocknum, 2),
		util.NewInt(0),
	)
	time.Sleep(3 * time.Second)

	exitId := LastExitId(plasma, userAddress)
	ChallengeExit(
		plasma,
		privateKeyECDSA,
		userAddress,
		txs,
		merkle,
		// Challenge exit of deposit with tx1 of new block.
		util.Sub(blocknum, 1),
		util.NewInt(0),
		exitId,
	)
	time.Sleep(3 * time.Second)
}

func finalize(
	plasma *contracts.Plasma,
	privateKeyECDSA *ecdsa.PrivateKey,
	userAddress string,
) {
	// Current block is the next block. Should we change that?
	blocknum := CurrentChildBlock(plasma, userAddress)
	txs := createSubmitBlockTxs(blocknum, userAddress)
	merkle := CreateMerkleTree(txs)

	SubmitBlock(plasma, privateKeyECDSA, userAddress, txs, merkle)
	time.Sleep(5 * time.Second)

	blocknum = CurrentChildBlock(plasma, userAddress)
	StartExit(
		plasma,
		privateKeyECDSA,
		userAddress,
		txs,
		merkle,
		util.Sub(blocknum, 1),
		util.NewInt(2),
	)
	time.Sleep(3 * time.Second)

	Finalize(plasma, privateKeyECDSA)
}

func logs(plasma *contracts.Plasma) {
	DepositFilter(plasma)
	time.Sleep(3 * time.Second)
	SubmitBlockFilter(plasma)
	time.Sleep(3 * time.Second)
	ExitStartedFilter(plasma)
	time.Sleep(3 * time.Second)
	DebugAddressFilter(plasma)
	time.Sleep(3 * time.Second)
	DebugUintFilter(plasma)
	time.Sleep(3 * time.Second)
	DebugBoolFilter(plasma)
	time.Sleep(3 * time.Second)
	DebugBytes32Filter(plasma)
	time.Sleep(3 * time.Second)
	DebugBytesFilter(plasma)
	time.Sleep(3 * time.Second)
	ChallengeSuccessFilter(plasma)
	time.Sleep(3 * time.Second)
	ChallengeFailureFilter(plasma)
	time.Sleep(3 * time.Second)
	FinalizeExitFilter(plasma)
	time.Sleep(3 * time.Second)
}

func createDepositTx(userAddress string, value int) chain.Transaction {
	return createTestTransaction(
		chain.ZeroInput(),
		&chain.Output{
			NewOwner: common.HexToAddress(userAddress),
			Amount:   util.NewInt(value),
		},
	)
}

func createSubmitBlockTxs(blocknum *big.Int, userAddress string) []chain.Transaction {
	return []chain.Transaction{
		createTestTransaction(
			&chain.Input{
				// Reference deposit
				BlkNum: util.Sub(blocknum, 1).Uint64(),
				TxIdx:  0,
				OutIdx: 0,
			},
			&chain.Output{
				NewOwner: common.HexToAddress(userAddress),
				Amount:   util.NewInt64(100),
			},
		), createTestTransaction(
			&chain.Input{
				// Reference current block
				BlkNum: blocknum.Uint64(),
				TxIdx:  0,
				OutIdx: 0,
			},
			&chain.Output{
				NewOwner: common.HexToAddress(userAddress),
				Amount:   util.NewInt64(200),
			},
		), createTestTransaction(
			&chain.Input{
				BlkNum: blocknum.Uint64(),
				TxIdx:  1, // Reference the previous transaction.
				OutIdx: 0,
			},
			&chain.Output{
				NewOwner: common.HexToAddress(userAddress),
				Amount:   util.NewInt64(300),
			},
		),
	}
}

func createTestTransaction(
	input0 *chain.Input,
	output0 *chain.Output,
) chain.Transaction {
	return chain.Transaction{
		Input0:  input0,
		Input1:  chain.ZeroInput(),
		Sig0:    []byte{},
		Sig1:    []byte{},
		Output0: output0,
		Output1: chain.ZeroOutput(),
		Fee:     new(big.Int),
		BlkNum:  uint64(0),
		TxIdx:   0,
	}
}

func exists(str string) bool {
	return len(str) > 0
}

// TODO: test with two inputs and outputs
// func exitAndChallengeTwoInputsOutputs

// TODO: test with multiple users
// func exitAndChallengeTwoUsers

// TODO: need to determine if transactions are successfull usually.
// func verifyTxSuccess(tx *types.Transaction) {}
