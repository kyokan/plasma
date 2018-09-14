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
	"log"
)

func Start(config *config.GlobalConfig, privateKey *ecdsa.PrivateKey) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	plasma, err := eth.NewClient(config.NodeURL, config.ContractAddr, privateKey)
	if err != nil {
		return err
	}

	ldb, storage, err := db.CreateStorage(path.Join(config.DBPath, "root"), plasma)
	if err != nil {
		return err
	}
	defer ldb.Close()

	sink := node.NewTransactionSink(storage)
	p := node.NewPlasmaNode(storage, sink, plasma)
	go p.Start()
	// TODO: ensure that 1 deposit tx is always 1 block
	go node.StartDepositListener(storage, sink, plasma)

	server := NewServer(ctx, storage, sink)
	go server.Start(config.RPCPort)
	log.Printf("Started RPC server on port %d", config.RPCPort)

	// TODO: add an exit listener to make sure to add an exit transaction to root node.
	// Also add an exit block to the plasma contract.

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<- c
	cancel()
	return nil
}