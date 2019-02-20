package db

import (
	"github.com/ethereum/go-ethereum/rlp"
)

type ExitLocator struct {
	PlasmaBlockNumber       uint64
	PlasmaTransactionIndex  uint32
	PlasmaOutputIndex       uint8
	EthereumBlockNumber     uint64
	EthereumTransactionHash []byte
}

func (e *ExitLocator) MarshalBinary() ([]byte, error) {
	return rlp.EncodeToBytes(e)
}

func (e *ExitLocator) UnmarshalBinary(data []byte) error {
	return rlp.DecodeBytes(data, e)
}
