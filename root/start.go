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

        server, cancel := BuildServer(config.RPCPort, config.NodeURL, config.ContractAddr, config.DBPath, privateKey)
        if server == nil {
          panic("no server")
        }
        defer cancel()

        fmt.Println("Started...\n")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
        fmt.Println("Received an interrupt, stopping services...\n")
	return nil
}

func BuildServer(rpcPort int, nodeURL string, contractAddr string, dbPath string, privateKey *ecdsa.PrivateKey) (*Server, func()) {
	ethClient, err := eth.NewClient(nodeURL, contractAddr, privateKey)
	if err != nil {
          panic(err)
	}

	ldb, storage, err := db.CreateStorage(path.Join(dbPath, "root"))
	if err != nil {
          panic(err)
	}

	mpool := node.NewMempool(storage, ethClient)
	err = mpool.Start()
	if err != nil {
          panic(err)
	}

	chainsaw := node.NewChainsaw(ethClient, mpool, storage)
	if err := chainsaw.Start(); err != nil {
          panic(err)
	}

	confirmer := node.NewTransactionConfirmer(storage, ethClient)
	submitter := node.NewBlockSubmitter(ethClient, storage)
	if err := submitter.Start(); err != nil {
          panic(err)
	}

	p := node.NewPlasmaNode(storage, mpool, ethClient, submitter)
	go p.Start()

	ctx, cancel := context.WithCancel(context.Background())
	server := NewServer(ctx, storage, mpool, confirmer)
	go server.Start(rpcPort)

        return server, func() {
          ldb.Close()
          cancel()
        }
}
