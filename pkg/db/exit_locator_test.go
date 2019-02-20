package db

import (
	"testing"
	"github.com/kyokan/plasma/util"
	"github.com/stretchr/testify/require"
	"bytes"
)

func TestExitLocator_MarshalUnmarshal(t *testing.T) {
	hash1 := util.Sha256([]byte("hash1"))

	locator := &ExitLocator{
		EthereumBlockNumber:     999,
		EthereumTransactionHash: hash1,
		PlasmaBlockNumber:   2,
		PlasmaTransactionIndex: 6,
		PlasmaOutputIndex:       1,
	}
	locBytes, err := locator.MarshalBinary()
	require.NoError(t, err)
	var otherLocator ExitLocator
	err = otherLocator.UnmarshalBinary(locBytes)
	require.NoError(t, err)
	require.Equal(t, locator.EthereumBlockNumber, otherLocator.EthereumBlockNumber)
	require.True(t, bytes.Equal(locator.EthereumTransactionHash, otherLocator.EthereumTransactionHash))
	require.Equal(t, locator.PlasmaBlockNumber, otherLocator.PlasmaBlockNumber)
	require.Equal(t, locator.PlasmaTransactionIndex, otherLocator.PlasmaTransactionIndex)
	require.Equal(t, locator.PlasmaOutputIndex, otherLocator.PlasmaOutputIndex)
}
