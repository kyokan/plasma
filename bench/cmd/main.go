package main

import (
	"fmt"
	"github.com/kyokan/plasma/bench"
	"os"
	"runtime/trace"
	"time"
)

func main() {
	f, err := os.Create(time.Now().Format("bench-trace-2006-01-02T150405.pprof"))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := trace.Start(f); err != nil {
		panic(err)
	}
	defer trace.Stop()

	fmt.Println("Running benchmarker. This will take a while.")
	res, err := bench.BenchmarkSend(1000, 1)
	if err != nil {
		fmt.Println("benchmarking failed", err)
		os.Exit(1)
	}

	fmt.Println("TPS:", res.TPS)
	fmt.Println("Elapsed Time:", res.ElapsedTime.String())
	fmt.Println("Avg. Runtime:", res.AvgRunTime)
	fmt.Println("Completed Transactions:", res.CompletedTransactions)
	fmt.Println("Failed Transactions:", res.FailedTransactions)
}
