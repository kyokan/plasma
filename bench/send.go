package bench

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	harness "github.com/kyokan/plasma/cmd/harness/cmd"
	plasmacli "github.com/kyokan/plasma/cmd/plasmacli/cmd"
	"github.com/kyokan/plasma/pkg/eth"
	"github.com/kyokan/plasma/pkg/rpc/pb"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

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

func startPlasma(dbPath string) (*exec.Cmd, error) {
	plasma := exec.Command(
		path.Join(getRepoBase(), "target", "plasmad"),
		"--node-url",
		"http://localhost:8545",
		"--contract-addr",
		"0xF12b5dd4EAD5F743C6BaA640B0216200e89B60Da",
		"--private-key",
		"c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3",
		"--db",
		dbPath,
		"start-root",
	)
	if err := harness.LogCmd(plasma, "plasma"); err != nil {
		return nil, err
	}
	if err := plasma.Start(); err != nil {
		return nil, err
	}

	return plasma, nil
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

	plasma, err := startPlasma(plasmaDbPath)
	if err != nil {
		return nil, err
	}

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
				}

				if err := plasmacli.SpendDeposit(plasmaClient, account.client, account.priv, account.addr, zeroAddr, big.NewInt(1), receipt.DepositNonce); err != nil {
					fmt.Println("failed to spend deposit", err)
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
		// send interupt because we want the daemon to save the trace
		if err := plasma.Process.Signal(os.Interrupt); err != nil {
			fmt.Println("failed to stop plasma", err)
		}
		if err := os.RemoveAll(ganacheDbPath); err != nil {
			fmt.Println("failed to remove ganache DB")
		}
		if err := os.RemoveAll(plasmaDbPath); err != nil {
			fmt.Println("failed to remove plasma db")
		}
	}, nil
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

				if err := plasmacli.SpendTx(plasmaClient, account.priv, account.addr, zeroAddr, big.NewInt(1)); err != nil {
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
		if min > rt {
			min = rt
		}
		if max < rt {
			max = rt
		}
		sumRuntime += rt
	}

	avgRuntime := (float64(sumRuntime) / float64(completionCount)) / 1000000
	tps := (float64(completionCount) / float64(elapsed.Nanoseconds())) * 1000000000

	fmt.Println("Min Time:", min/1000000)
	fmt.Println("Max Time:", max/1000000)

	return &SendBenchmarkResult{
		ElapsedTime:           elapsed,
		AvgRunTime:            avgRuntime,
		CompletedTransactions: completionCount,
		FailedTransactions:    failureCount,
		TPS:                   tps,
	}, nil
}
