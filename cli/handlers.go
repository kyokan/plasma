package cli

import (
	"encoding/json"
	"fmt"
	"github.com/kyokan/plasma/merkle"
	"github.com/kyokan/plasma/txdag"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/eth"
	"context"
	"github.com/kyokan/plasma/config"
	"crypto/ecdsa"
	"google.golang.org/grpc"
	"github.com/kyokan/plasma/rpc/pb"
	"github.com/kyokan/plasma/rpc"
	"github.com/pkg/errors"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/olekukonko/tablewriter"
	"os"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"strconv"
)

func initHandler(config *config.GlobalConfig, privateKey *ecdsa.PrivateKey, pCtx context.Context) (eth.Client, error) {
	client, err := eth.NewClient(config.NodeURL, config.ContractAddr, privateKey)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func Finalize(config *config.GlobalConfig, privateKey *ecdsa.PrivateKey) error {
	client, err := initHandler(config, privateKey, context.Background())
	if err != nil {
		return err
	}

	err = client.FinalizeDepositExits()
	if err != nil {
		return err
	}
	return client.FinalizeTransactionExits()
}

func Exit(config *config.GlobalConfig, privateKey *ecdsa.PrivateKey, rootHost string, blockNum uint64, txIndex uint32, oIndex uint8) error {
	ctx := context.Background()
	client, err := initHandler(config, privateKey, ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Exit starting for blocknum: %d, txIndex: %d, oIndex: %d\n", blockNum, txIndex, oIndex)

	conn, err := grpc.Dial(rootHost)
	if err != nil {
		return err
	}
	defer conn.Close()

	rc := pb.NewRootClient(conn)
	res, err := rc.GetBlock(ctx, &pb.GetBlockRequest{
		Number: blockNum,
	})
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("block does not exist")
	}
	transactions := rpc.DeserializeTxs(res.Transactions)
	hashables := make([]merkle.DualHashable, len(transactions))
	for i := 0; i < len(transactions); i++ {
		hashables[i] = &transactions[i]
	}
	proof, err := merkle.GetProof(hashables, merkle.DefaultDepth, int32(txIndex))
	transaction := transactions[txIndex]
	opts := &eth.StartExitOpts{
		Transaction: transaction,
		Input: *chain.NewInputFromTransaction(transaction, int64(oIndex)),
		Signature: []byte{},
		ConfirmSignature: []byte{},
		Proof: proof,
	}

	return client.StartTransactionExit(opts)
}

func Deposit(config *config.GlobalConfig, privateKey *ecdsa.PrivateKey, amount *big.Int) error {
	ctx := context.Background()
	client, err := initHandler(config, privateKey, ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Deposit starting for amount: %s\n", amount.Text(10))
	userAddress := crypto.PubkeyToAddress(*(privateKey.Public()).(*ecdsa.PublicKey))
	t := createDepositTx(userAddress, amount)
	err = client.Deposit(amount, &t)
	time.Sleep(3 * time.Second)
	return err
}

func Balance(rootHost string, address common.Address) error {
	log.Info("Received Balance request.")

	conn, err := grpc.Dial(rootHost, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	rc := pb.NewRootClient(conn)
	res, err := rc.GetBalance(context.Background(), &pb.GetBalanceRequest{
		Address: address.Bytes(),
	})
	if err != nil {
		return err
	}

	fmt.Printf("Balance: %s\n", rpc.DeserializeBig(res.Balance).Text(10))
	return nil
}

func Block(rootHost string, blockNum uint64) error {
	log.Info("Received Block request.")

	conn, err := grpc.Dial(rootHost, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	rc := pb.NewRootClient(conn)
	res, err := rc.GetBlock(context.Background(), &pb.GetBlockRequest{
		Number: blockNum,
	})
	if err != nil {
		return err
	}

	block := rpc.DeserializeBlock(res.Block)
	blockTable := tablewriter.NewWriter(os.Stdout)
	blockTable.SetHeader([]string{"Hash", "MerkleRoot", "RLPMerkleRoot", "PrevHash", "Number"})
	blockTable.Append([]string{
		hexutil.Encode(block.BlockHash),
		hexutil.Encode(block.Header.MerkleRoot),
		hexutil.Encode(block.Header.RLPMerkleRoot),
		hexutil.Encode(block.Header.PrevHash),
		strconv.FormatUint(block.Header.Number, 10),
	})

	txs := rpc.DeserializeTxs(res.Transactions)
	txsTable := txsTable(txs)

	blockTable.Render()
	fmt.Print("\n")
	txsTable.Render()

	return nil
}

func UTXOs(rootHost string, address common.Address) error {
	log.Info("Received UTXOs request.")

	conn, err := grpc.Dial(rootHost, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	rc := pb.NewRootClient(conn)
	res, err := rc.GetOutputs(context.Background(), &pb.GetOutputsRequest{
		Address: address.Bytes(),
	})
	if err != nil {
		return err
	}

	txs := rpc.DeserializeTxs(res.Transactions)
	table := txsTable(txs)
	table.Render()
	return nil
}

func Send(privateKey *ecdsa.PrivateKey, rootHost string, from, to common.Address, amount *big.Int) error {
	ctx := context.Background()
	conn, err := grpc.Dial(rootHost, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	rc := pb.NewRootClient(conn)

	utxoResponse, err := rc.GetOutputs(ctx, &pb.GetOutputsRequest{
		Address:   from.Bytes(),
		Spendable: true,
	})
	if err != nil {
		return err
	}
	utxos := rpc.DeserializeTxs(utxoResponse.Transactions)
	tx, err := txdag.FindBestUTXOs(from, to, amount, utxos)
	if err != nil {
		return err
	}
	j, _ := json.MarshalIndent(&tx, "", "\t")
	fmt.Println(string(j))

	hash := eth.GethHash(tx.SignatureHash())
	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return err
	}

	copy(tx.Sig0[:], signature)
	if !tx.Input1.IsZeroInput() {
		copy(tx.Sig1[:], signature)
	}

	res, err := rc.Send(ctx, &pb.SendRequest{
		Transaction: rpc.SerializeTx(tx),
	})
	if err != nil {
		return err
	}
	tx = rpc.DeserializeTx(res.Transaction)
	jsonTx, err := json.MarshalIndent(&tx, "", "\t")
	fmt.Printf("Send results: %s", string(jsonTx))
	return nil
}

// TODO: Use same code as transaction sink.
func createDepositTx(userAddress common.Address, value *big.Int) chain.Transaction {
	return chain.Transaction{
		Input0: chain.ZeroInput(),
		Input1: chain.ZeroInput(),
		Output0: &chain.Output{
			Owner: userAddress,
			Denom: value,
		},
		Output1: chain.ZeroOutput(),
		Fee:     big.NewInt(0),
	}
}

func txsTable(txs []chain.Transaction) *tablewriter.Table {
	txsTable := tablewriter.NewWriter(os.Stdout)
	txsTable.SetHeader([]string{
		"Idx",
		"Input0Block",
		"Input0TxIdx",
		"Input0OutIdx",
		"Input0Sig",
		"Input1Block",
		"Input1TxIdx",
		"Input1OutIdx",
		"Input1Sig",
		"Output0Owner",
		"Output0Amount",
		"Output1Owner",
		"Output1Amount",
		"Fee",
	})
	for _, tx := range txs {
		txsTable.Append([]string{
			tx.TxIdx.String(),
			tx.Input0.BlkNum.String(),
			tx.Input0.TxIdx.String(),
			tx.Input0.OutIdx.String(),
			hexutil.Encode(tx.Sig0[:]),
			tx.Input1.BlkNum.String(),
			tx.Input1.TxIdx.String(),
			tx.Input1.OutIdx.String(),
			hexutil.Encode(tx.Sig1[:]),
			tx.Output0.Owner.Hex(),
			tx.Output0.Denom.Text(10),
			tx.Output1.Owner.Hex(),
			tx.Output1.Denom.Text(10),
			tx.Fee.Text(10),
		})
	}
	return txsTable
}
