package main

import (
	"fmt"
	"github.com/kyokan/plasma/bench"
	"os"
)

func main() {
	fmt.Println("Running benchmarker. This will take a while.")
	res, err := bench.BenchmarkSend100()
	if err != nil {
		fmt.Println("benchmarking failed", err)
		os.Exit(1)
	}

	fmt.Println("TPS:", res.TPS)
	fmt.Println("Elapsed Time:", res.ElapsedTime.String())
	fmt.Println("Completed Transactions:", res.CompletedTransactions)
	fmt.Println("Failed Transactions:", res.FailedTransactions)
}
