package cmd

import (
	"github.com/spf13/cobra"
	"crypto/ecdsa"
	"github.com/pkg/errors"
	"io/ioutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/kyokan/plasma/pkg/rpc/pb"
	"google.golang.org/grpc"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"strings"
	"github.com/ethereum/go-ethereum/common"
)

func AddrOrPrivateKeyAddr(cmd *cobra.Command, args []string, addrArg int) (common.Address, error) {
	var addr common.Address
	if len(args) > addrArg {
		addr = common.HexToAddress(args[addrArg])
	} else {
		privKey, err := ParsePrivateKey(cmd)
		if err != nil {
			return addr, err
		}

		addr = crypto.PubkeyToAddress(privKey.PublicKey)
	}

	return addr, nil
}

func ParsePrivateKey(cmd *cobra.Command) (*ecdsa.PrivateKey, error) {
	path := cmd.Flag(FlagPrivateKeyPath).Value.String()
	if path == "" {
		return nil, errors.New("no private key path set")
	}

	expanded, err := homedir.Expand(path)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't expand homedir")
	}

	keyBytes, err := ioutil.ReadFile(expanded)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read private key")
	}

	return crypto.HexToECDSA(strings.TrimSpace(string(keyBytes)))
}

func CreateRootClient(url string) (pb.RootClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to dial node")
	}

	client := pb.NewRootClient(conn)
	return client, conn, nil
}

func PrintJSON(in interface{}) error {
	j, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(j))
	return nil
}