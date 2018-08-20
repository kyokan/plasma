package validator

import (
	"fmt"
	"log"
	"path"

	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
	"github.com/kyokan/plasma/config"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/grpc"
	"github.com/kyokan/plasma/rpc/pb"
	"context"
)

func Start(config *config.GlobalConfig, privateKey *ecdsa.PrivateKey, rootHost string) error {
	log.Println("Validator Starting")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	userAddress := crypto.PubkeyToAddress((privateKey.Public()).(ecdsa.PublicKey))
	ethClient, err := eth.NewClient(config.NodeURL, config.ContractAddr, privateKey)
	if err != nil {
		return err
	}

	ldb, storage, err := db.CreateStorage(path.Join(config.DBPath, "validator", userAddress.Hex()), ethClient)
	if err != nil {
		return err
	}
	defer ldb.Close()

	conn, err := grpc.Dial(fmt.Sprintf("http://%s/rpc", rootHost))
	if err != nil {
		return err
	}
	defer conn.Close()

	rootClient := pb.NewRootClient(conn)
	go RootNodeListener(ctx, storage, ethClient, rootClient, userAddress)
	go ExitStartedListener(ctx, storage, ethClient, rootClient)
	go Run(config.RPCPort)

	select {}
	return nil
}
