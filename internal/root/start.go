package root

import (
	"github.com/kyokan/plasma/pkg/eth"
	"github.com/kyokan/plasma/pkg/db"
	"github.com/kyokan/plasma/pkg/config"
	"crypto/ecdsa"
	"github.com/kyokan/plasma/pkg/service"
	"context"
	"os"
	"os/signal"
	"path"
)

func Start(config *config.GlobalConfig, privateKey *ecdsa.PrivateKey) error {
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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	return nil
}
