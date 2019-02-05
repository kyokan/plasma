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

	p := node.NewPlasmaNode(storage, mpool, plasma)
	go p.Start()
	// TODO: ensure that 1 deposit tx is always 1 block

	go node.StartExitListener(storage, plasma, context.Background())

	server := NewServer(ctx, storage, mpool, confirmer)
	go server.Start(config.RPCPort)

	// TODO: add an exit listener to make sure to add an exit transaction to root node.
	// Also add an exit block to the plasma contract.

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	cancel()
	return nil
}
