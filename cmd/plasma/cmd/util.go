package cmd

import (
	"github.com/kyokan/plasma/config"
	"github.com/spf13/viper"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

func NewGlobalConfig() *config.GlobalConfig {
	return &config.GlobalConfig{
		DBPath:       viper.GetString(FlagDB),
		NodeURL:      viper.GetString(FlagNodeURL),
		RPCPort:      viper.GetInt(FlagRPCPort),
		ContractAddr: viper.GetString(FlagContractAddr),
	}
}

func ParsePrivateKey() (*ecdsa.PrivateKey, error) {
	privateKeyStr := viper.GetString(FlagPrivateKey)
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse private key")
	}
	return privateKey, nil
}