package root

import (
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/config"
	"crypto/ecdsa"
	"github.com/kyokan/plasma/node"
	"context"
	"os"
	"os/signal"
	"path"
        "runtime/trace"
        "time"
        "fmt"
)

func Start(config *config.GlobalConfig, privateKey *ecdsa.PrivateKey) error {
        f, err := os.Create(time.Now().Format("daemon-trace-2006-01-02T150405.pprof"))
        if err != nil {
          panic(err)
        }
        defer f.Close()

        if err := trace.Start(f); err != nil {
          panic(err)
        }
        defer trace.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ethClient, err := eth.NewClient(config.NodeURL, config.ContractAddr, privateKey)
	if err != nil {
		return err
	}

	ldb, storage, err := db.CreateStorage(path.Join(config.DBPath, "root"))
	if err != nil {
		return err
	}
	defer ldb.Close()

	mpool := node.NewMempool(storage, ethClient)
	err = mpool.Start()
	if err != nil {
		return err
	}

	chainsaw := node.NewChainsaw(ethClient, mpool, storage)
	if err := chainsaw.Start(); err != nil {
	    return err
	}

	confirmer := node.NewTransactionConfirmer(storage, ethClient)
	submitter := node.NewBlockSubmitter(ethClient, storage)
	if err := submitter.Start(); err != nil {
	    return err
	}

	p := node.NewPlasmaNode(storage, mpool, ethClient, submitter)
	go p.Start()

	server := NewServer(ctx, storage, mpool, confirmer)
	go server.Start(config.RPCPort)

        fmt.Println("Started...\n")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
        fmt.Println("Received an interrupt, stopping services...\n")
	cancel()
	return nil
}
