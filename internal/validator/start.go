package validator

import (
	"github.com/kyokan/plasma/pkg/config"
	"crypto/ecdsa"
	"github.com/kyokan/plasma/pkg/eth"
	"github.com/kyokan/plasma/pkg/db"
	"github.com/kyokan/plasma/pkg/service"
	"os"
	"os/signal"
	"path"
	"github.com/kyokan/plasma/pkg/rpc/pb"
	"google.golang.org/grpc"
	"github.com/kyokan/plasma/pkg/log"
	"context"
	)

var valStartLogger = log.ForSubsystem("ValidatorStart")

func Start(config *config.GlobalConfig, rootUrl string, privateKey *ecdsa.PrivateKey) error {
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

	conn, err := grpc.Dial(rootUrl, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	rootClient := pb.NewRootClient(conn)

	syncer := service.NewSyncer(storage, rootClient, ethClient)
	if err := syncer.Start(); err != nil {
		return err
	}

	server := NewServer(ctx, storage, rootClient)
	go func() {
		if err := server.Start(config.RPCPort); err != nil {
		    log.WithError(valStartLogger, err).Error("failed to start server")
		    os.Exit(1)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	if err := syncer.Stop(); err != nil {
		log.WithError(valStartLogger, err).Error("failed to stop syncer")
	}
	return nil
}
