package root

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/kyokan/plasma/pkg/config"
	"github.com/kyokan/plasma/pkg/db"
	"github.com/kyokan/plasma/pkg/eth"
	"github.com/kyokan/plasma/pkg/service"
	"os"
	"os/signal"
	"path"
	"runtime/trace"
	"time"
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

	ldb, storage, err := db.CreateLevelStorage(path.Join(config.DBPath, "root"))
	if err != nil {
		return err
	}
	defer ldb.Close()

	mpool := service.NewMempool(storage, ethClient)
	err = mpool.Start()
	if err != nil {
		return err
	}

	chainsaw := service.NewChainsaw(ethClient, mpool, storage)
	if err := chainsaw.Start(); err != nil {
		return err
	}

	confirmer := service.NewTransactionConfirmer(storage, ethClient)
	submitter := service.NewBlockSubmitter(ethClient, storage)
	if err := submitter.Start(); err != nil {
		return err
	}

	p := service.NewPlasmaNode(storage, mpool, ethClient, submitter)
	go p.Start()

	server := NewServer(ctx, storage, mpool, confirmer)
	go server.Start(config.RPCPort)

	rest := NewRESTServer(storage, mpool, confirmer, 6546)
	go rest.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("Received an interrupt, stopping services...")
	return nil
}
