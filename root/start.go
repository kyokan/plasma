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
)

func Start(config *config.GlobalConfig, privateKey *ecdsa.PrivateKey) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	plasma, err := eth.NewClient(config.NodeURL, config.ContractAddr, privateKey)
	if err != nil {
		return err
	}

	ldb, storage, err := db.CreateStorage(path.Join(config.DBPath, "root"))
	if err != nil {
		return err
	}
	defer ldb.Close()

	mpool := node.NewMempool(storage)
	err = mpool.Start()
	if err != nil {
		return err
	}

	chainsaw := node.NewChainsaw(plasma, mpool, storage)
	if err := chainsaw.Start(); err != nil {
	    return err
	}

	confirmer := node.NewTransactionConfirmer(storage)
	submitter := node.NewBlockSubmitter(plasma, storage)
	if err := submitter.Start(); err != nil {
	    return err
	}

	p := node.NewPlasmaNode(storage, mpool, plasma, submitter)
	go p.Start()

	server := NewServer(ctx, storage, mpool, confirmer)
	go server.Start(config.RPCPort)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	cancel()
	return nil
}
