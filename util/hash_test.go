package util

import (
	"github.com/stretchr/testify/suite"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
	"fmt"
	"encoding/json"
	"testing"
)

type testHashStruct struct {
	Hash Hash
}

var hashStr = "0x495af4c7f191a5aaf437b403d57ae277f72c626afabf81716e9711d848b7c6e4"

type hashSuite struct {
	suite.Suite
	hash Hash
}

func (h *hashSuite) SetupTest() {
	hashData, err := hexutil.Decode(hashStr)
	require.NoError(h.T(), err)
	var hash [32]byte
	copy(hash[:], hashData)
	h.hash = hash[:]
}

func (h *hashSuite) TestHash_MarshalJSON() {
	data, err := json.Marshal(h.hash)
	require.NoError(h.T(), err)
	require.Equal(h.T(), fmt.Sprintf("\"%s\"", hashStr), string(data))
}

func (h *hashSuite) TestHash_UnmarshalJSON() {
	var testSig Hash
	jsonStr := fmt.Sprintf("\"%s\"", hashStr)
	json.Unmarshal([]byte(jsonStr), &testSig)
	require.Equal(h.T(), h.hash, testSig)
}

func (h *hashSuite) TestHash_MarshallingInStruct() {
	testStruct := testHashStruct{
		Hash: h.hash,
	}
	data, err := json.Marshal(testStruct)
	require.NoError(h.T(), err)
	var unmarshalledTestStruct testHashStruct
	err = json.Unmarshal(data, &unmarshalledTestStruct)
	require.Equal(h.T(), testStruct.Hash, unmarshalledTestStruct.Hash)
}

func TestHashSuite(t *testing.T) {
	suite.Run(t, new(hashSuite))
}