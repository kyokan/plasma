package bench

import (
	harness "github.com/kyokan/plasma/cmd/harness/cmd"
	plasmacli "github.com/kyokan/plasma/cmd/plasmacli/cmd"
	plasmaRoot "github.com/kyokan/plasma/root"
	"github.com/kyokan/plasma/chain"
	"io/ioutil"
	"fmt"
	"os"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/kyokan/plasma/eth"
	"math/big"
	"sync"
	"github.com/ethereum/go-ethereum/common"
	"strings"
        "math"
	"github.com/kyokan/plasma/rpc/pb"
	"time"
	"sync/atomic"
        "errors"
	"bytes"
        "sort"
	"context"
	"github.com/kyokan/plasma/util"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type sendCmdOutput struct {
	Value            string   `json:"value"`
	To               string   `json:"to"`
	BlockNumber      uint64   `json:"blockNumber"`
	TransactionIndex uint32   `json:"transactionIndex"`
	DepositNonce     string   `json:"depositNonce"`
	MerkleRoot       string   `json:"merkleRoot"`
	ConfirmSigs      []string `json:"confirmSigs"`
}

type StopFunc func()

type SendBenchmarkResult struct {
	ElapsedTime           time.Duration
        AvgRunTime            float64
	CompletedTransactions int64
	FailedTransactions    int64
	TPS                   float64
}

type benchAccount struct {
	client eth.Client
	priv   *ecdsa.PrivateKey
	addr   common.Address
}

var accounts []*benchAccount
var plasmaClient pb.RootClient
var zeroAddr common.Address
var plasmaDaemonServer *plasmaRoot.Server

func getRepoBase() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	splits := strings.Split(dir, string(os.PathSeparator))

	for i := len(splits) - 1; i >= 0; i-- {
		if splits[i] == "plasma" {
			repoBase := splits[0 : i+1]
			return strings.Join(repoBase, string(os.PathSeparator))
		}
	}

	panic("could not determine repo base")
}

func startPlasmaRoot(dbPath string) context.CancelFunc {
  privateKey, err := crypto.HexToECDSA("c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3")
  if err != nil {
    panic(err)
  }
  pds, cancel := plasmaRoot.BuildServer(
    6545,
    "http://localhost:8545",
    "0xF12b5dd4EAD5F743C6BaA640B0216200e89B60Da",
    dbPath,
    privateKey,
  )
  plasmaDaemonServer = pds
  time.Sleep(5 * time.Second)
  return cancel
}

func initSendBench(accountCount int) (StopFunc, error) {
	ganacheDbPath, err := ioutil.TempDir("", "ganache")
	if err != nil {
		return nil, err
	}
	plasmaDbPath, err := ioutil.TempDir("", "plasma")
	if err != nil {
		return nil, err
	}
	ganache, err := harness.StartGanache(8545, 1, accountCount, ganacheDbPath)
	if err != nil {
		return nil, err
	}
	harness.MigrateGanache(getRepoBase())

	var ethClients []eth.Client
        for _, privStr := range benchPrivateKeys[0:accountCount] {
		priv, err := crypto.HexToECDSA(privStr)
		if err != nil {
			return nil, err
		}
		addr := crypto.PubkeyToAddress(priv.PublicKey)

		client, err := eth.NewClient("http://localhost:8545", "0xF12b5dd4EAD5F743C6BaA640B0216200e89B60Da", priv)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, &benchAccount{
			client: client,
			priv:   priv,
			addr:   addr,
		})
	}

        cancel := startPlasmaRoot(plasmaDbPath)
	pClient, conn, err := plasmacli.CreateRootClient("localhost:6545")
	if err != nil {
		return nil, err
	}
	plasmaClient = pClient

	var pendingAccounts []*benchAccount
	depositConcurrency := 10
	for i, account := range accounts {
		pendingAccounts = append(pendingAccounts, account)
		if len(pendingAccounts) < depositConcurrency && i != len(ethClients)-1 {
			continue
		}

		var wg sync.WaitGroup
		wg.Add(len(pendingAccounts))
		for _, client := range pendingAccounts {
			go func(account *benchAccount) {
				defer wg.Done()
				val := big.NewInt(1000000)
				receipt, err := plasmacli.Deposit(account.client, val)
				if err != nil {
					fmt.Println("failed to deposit", err)
				} else {
                                  fmt.Printf("Completed deposit...")
                                }

				if err := plasmacli.SpendDeposit(plasmaClient, account.client, account.priv, account.addr, zeroAddr, big.NewInt(1), receipt.DepositNonce); err != nil {
					fmt.Println("failed to spend deposit", err)
				} else {
                                  fmt.Printf("Completed deposit spend...")
                                }
			}(client)
		}
		wg.Wait()
		pendingAccounts = make([]*benchAccount, 0)
		fmt.Printf("Completed %d of %d deposits...\n", i+1, len(accounts))
	}

	return func() {
		if err := conn.Close(); err != nil {
			fmt.Println("failed to close gRPC client connection")
		}
		if err := ganache.Process.Kill(); err != nil {
			fmt.Println("failed to stop ganache", err)
		}
                cancel()
		if err := os.RemoveAll(ganacheDbPath); err != nil {
			fmt.Println("failed to remove ganache DB")
		}
		if err := os.RemoveAll(plasmaDbPath); err != nil {
			fmt.Println("failed to remove plasma db")
		}
	}, nil
}

func selectUTXOs(confirmedTxs []chain.ConfirmedTransaction, addr common.Address, total *big.Int) ([]chain.ConfirmedTransaction, error) {
	sort.Slice(confirmedTxs, func(i, j int) bool {
		a := confirmedTxs[i].Transaction.Body.OutputFor(&addr).Amount
		b := confirmedTxs[j].Transaction.Body.OutputFor(&addr).Amount
		return a.Cmp(b) > 0
	})

	first := confirmedTxs[0]
	firstBody := first.Transaction.Body

	if firstBody.OutputFor(&addr).Amount.Cmp(total) >= 0 {
		return []chain.ConfirmedTransaction{first}, nil
	}

	for i := len(confirmedTxs) - 1; i >= 0; i-- {
		second := confirmedTxs[i]
		secondBody := second.Transaction.Body
		sum := big.NewInt(0)
		sum = sum.Add(sum, firstBody.OutputFor(&addr).Amount)
		sum = sum.Add(sum, secondBody.OutputFor(&addr).Amount)
		if sum.Cmp(total) >= 0 {
			return []chain.ConfirmedTransaction{
				first,
				second,
			}, nil
		}
	}

	return nil, errors.New("no suitable UTXOs found")
}

func SpendTx(client pb.RootClient, privKey *ecdsa.PrivateKey, from common.Address, to common.Address, value *big.Int) error {
	fmt.Println("selecting outputs")

        res, err := plasmaDaemonServer.GetOutputs(nil, &pb.GetOutputsRequest{
		Address:   from.Bytes(),
		Spendable: true,
	})
	if err != nil {
		return err
	}

	var utxos []chain.ConfirmedTransaction
	for _, utxoProto := range res.ConfirmedTransactions {
		confirmedTx, err := chain.ConfirmedTransactionFromProto(utxoProto)
		if err != nil {
			return err
		}
		utxos = append(utxos, *confirmedTx)
	}
	if len(utxos) == 0 {
		return errors.New("no spendable outputs")
	}
	selectedUTXOs, err := selectUTXOs(utxos, from, value)
	if err != nil {
		return err
	}

	total := big.NewInt(0)
	tx := &chain.Transaction{
		Body: chain.ZeroBody(),
	}
	for i, utxo := range selectedUTXOs {
		txBody := utxo.Transaction.Body
		var input *chain.Input
		if i == 0 {
			input = tx.Body.Input0
		} else if i == 1 {
			input = tx.Body.Input1
		} else {
			panic("too many inputs!")
		}

		input.BlockNum = txBody.BlockNumber
		input.TxIdx = txBody.TransactionIndex
		input.OutIdx = txBody.OutputIndexFor(&from)

		if i == 0 {
			tx.Body.Input0ConfirmSig = utxo.ConfirmSigs[0]
		} else {
			tx.Body.Input1ConfirmSig = utxo.ConfirmSigs[1]
		}

		total = total.Add(total, txBody.OutputFor(&from).Amount)
	}

	tx.Body.Output0.Amount = value
	tx.Body.Output0.Owner = to

	if total.Cmp(value) > 0 {
		totalClone := new(big.Int).Set(total)
		tx.Body.Output1.Amount = totalClone.Sub(totalClone, value)
		tx.Body.Output1.Owner = from
	}

	sig, err := eth.Sign(privKey, tx.Body.SignatureHash())
	if err != nil {
		return err
	}
	tx.Sigs[0] = sig
	tx.Sigs[1] = sig

	sendRes, err := plasmaDaemonServer.Send(nil, &pb.SendRequest{
		Transaction: tx.Proto(),
	})
	if err != nil {
		return err
	}

	tx.Body.BlockNumber = sendRes.Inclusion.BlockNumber
	tx.Body.TransactionIndex = sendRes.Inclusion.TransactionIndex
	var buf bytes.Buffer
	buf.Write(tx.RLPHash(util.Sha256))
	buf.Write(sendRes.Inclusion.MerkleRoot)
	sigHash := util.Sha256(buf.Bytes())
	confirmSig, err := eth.Sign(privKey, sigHash)
	if err != nil {
		return err
	}

	fmt.Println("confirming transaction")

	_, err = plasmaDaemonServer.Confirm(nil, &pb.ConfirmRequest{
		BlockNumber:      sendRes.Inclusion.BlockNumber,
		TransactionIndex: sendRes.Inclusion.TransactionIndex,
		ConfirmSig0:      confirmSig[:],
		ConfirmSig1:      confirmSig[:],
	})
	if err != nil {
		return err
	}

	out := &sendCmdOutput{
		Value:            value.Text(10),
		To:               to.Hex(),
		BlockNumber:      sendRes.Inclusion.BlockNumber,
		TransactionIndex: sendRes.Inclusion.TransactionIndex,
		MerkleRoot:       hexutil.Encode(sendRes.Inclusion.MerkleRoot),
		ConfirmSigs: []string{
			hexutil.Encode(confirmSig[:]),
			hexutil.Encode(confirmSig[:]),
		},
	}

	return plasmacli.PrintJSON(out)
}

func BenchmarkSend(accountCount int, benchCallMultiper int) (*SendBenchmarkResult, error) {
	stop, err := initSendBench(accountCount)
	if err != nil {
		return nil, err
	}

	start := time.Now()

        runtimes := make(chan int64, accountCount)
	failureCount := int64(0)
	completionCount := int64(0)

	for i := 0; i < benchCallMultiper; i++ {
		var wg sync.WaitGroup
		wg.Add(len(accounts))
		for _, account := range accounts {
			go func(account *benchAccount) {
				defer wg.Done()
                                transRuntime := time.Now()

                                if err := SpendTx(plasmaClient, account.priv, account.addr, zeroAddr, big.NewInt(1)); err != nil {
					atomic.AddInt64(&failureCount, 1)
					fmt.Println("failed to spend deposit", err)
				} else {
                                        transRuntimeElapsed := time.Since(transRuntime).Nanoseconds()
                                        runtimes <- transRuntimeElapsed
					atomic.AddInt64(&completionCount, 1)
				}
			}(account)
		}
		wg.Wait()
	}

	elapsed := time.Since(start)
	stop()

        close(runtimes)

        min := int64(math.MaxInt64)
        max := int64(math.MinInt64)
        sumRuntime := int64(0)
        for rt := range runtimes {
          if (min > rt) {
            min = rt
          }
          if (max < rt) {
            max = rt
          }
          sumRuntime += rt
        }

        avgRuntime := (float64(sumRuntime) / float64(completionCount)) / 1000000
	tps := (float64(completionCount) / float64(elapsed.Nanoseconds())) * 1000000000

        fmt.Println("Min Time:", min / 1000000)
        fmt.Println("Max Time:", max / 1000000)

	return &SendBenchmarkResult{
		ElapsedTime:           elapsed,
                AvgRunTime:            avgRuntime,
		CompletedTransactions: completionCount,
		FailedTransactions:    failureCount,
		TPS:                   tps,
	}, nil
}
