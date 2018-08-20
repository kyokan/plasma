package cli

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/eth"
	"context"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/config"
	"crypto/ecdsa"
	"path"
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

func initHandler(config *config.GlobalConfig, privateKey *ecdsa.PrivateKey, pCtx context.Context) (context.CancelFunc, eth.Client, db.PlasmaStorage, error) {
	ctx, cancel := context.WithCancel(pCtx)
	client, err := eth.NewClient(config.NodeURL, config.ContractAddr, privateKey)
	if err != nil {
		return cancel, nil, nil, err
	}

	ldb, storage, err := db.CreateStorage(path.Join(config.DBPath, "root"), client)
	if err != nil {
		return cancel, nil, nil, err
	}

	go func() {
		<-ctx.Done()
		ldb.Close()
	}()

	return cancel, client, storage, nil
}

func Finalize(config *config.GlobalConfig, privateKey *ecdsa.PrivateKey) error {
	cancel, client, _, err := initHandler(config, privateKey, context.Background())
	defer cancel()
	if err != nil {
		return err
	}

	return client.Finalize()
}

func Exit(config *config.GlobalConfig, privateKey *ecdsa.PrivateKey, rootHost string, blockNum *big.Int, txIndex uint, oIndex uint) error {
	ctx := context.Background()
	cancel, client, _, err := initHandler(config, privateKey, ctx)
	defer cancel()
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
		Number: rpc.SerializeBig(blockNum),
	})
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("block does not exist")
	}

	return client.StartExit(&eth.StartExitOpts{
		Block:    rpc.DeserializeBlock(res.Block),
		Txs:      rpc.DeserializeTxs(res.Transactions),
		BlockNum: blockNum,
		TxIndex:  txIndex,
		OutIndex: oIndex,
	})
}

func Deposit(config *config.GlobalConfig, privateKey *ecdsa.PrivateKey, amount *big.Int) error {
	ctx := context.Background()
	cancel, client, _, err := initHandler(config, privateKey, ctx)
	defer cancel()
	if err != nil {
		return err
	}

	fmt.Printf("Deposit starting for amount: %s\n", amount.Text(10))
	userAddress := crypto.PubkeyToAddress(*(privateKey.Public()).(*ecdsa.PublicKey))
	t := createDepositTx(userAddress, amount)
	client.Deposit(amount, &t)
	time.Sleep(3 * time.Second)
	curr, err := client.CurrentChildBlock()
	if err != nil {
		return err
	}

	fmt.Printf("Last child block: %v\n", curr)
	return nil
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

func Block(rootHost string, blockNum *big.Int) error {
	log.Info("Received Block request.")

	conn, err := grpc.Dial(rootHost, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	rc := pb.NewRootClient(conn)
	res, err := rc.GetBlock(context.Background(), &pb.GetBlockRequest{
		Number: rpc.SerializeBig(blockNum),
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
	res, err := rc.GetUTXOs(context.Background(), &pb.GetUTXOsRequest{
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

// TODO: Use same code as transaction sink.
func createDepositTx(userAddress common.Address, value *big.Int) chain.Transaction {
	return chain.Transaction{
		Input0: chain.ZeroInput(),
		Input1: chain.ZeroInput(),
		Output0: &chain.Output{
			NewOwner: userAddress,
			Amount:   value,
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
		txsTable.Append([]string {
			strconv.FormatUint(uint64(tx.TxIdx), 10),
			strconv.FormatUint(tx.Input0.BlkNum, 10),
			strconv.FormatUint(uint64(tx.Input0.TxIdx), 10),
			strconv.FormatUint(uint64(tx.Input0.OutIdx), 10),
			hexutil.Encode(tx.Sig0),
			strconv.FormatUint(tx.Input1.BlkNum, 10),
			strconv.FormatUint(uint64(tx.Input1.TxIdx), 10),
			strconv.FormatUint(uint64(tx.Input1.OutIdx), 10),
			hexutil.Encode(tx.Sig1),
			tx.Output0.NewOwner.Hex(),
			tx.Output0.Amount.Text(10),
			tx.Output1.NewOwner.Hex(),
			tx.Output1.Amount.Text(10),
			tx.Fee.Text(10),
		})
	}
	return txsTable
}