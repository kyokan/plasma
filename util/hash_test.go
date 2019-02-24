package util

import (
	"github.com/stretchr/testify/suite"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
	"fmt"
	"encoding/json"
	"testing"
	"bytes"
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

func TestKeccak256(t *testing.T) {
	expected, err := hexutil.Decode("0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8")
	require.NoError(t, err)
	data := []byte("hello")
	hash := Keccak256(data)
	require.True(t, bytes.Equal(expected, hash))
}

func TestSha256(t *testing.T) {
	expected, err := hexutil.Decode("0x2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824")
	require.NoError(t, err)
	data := []byte("hello")
	hash := Sha256(data)
	require.True(t, bytes.Equal(expected, hash))
}

func TestGethHash(t *testing.T) {
	expected, err := hexutil.Decode("0x31073a1d3eacb5cb2c0c310537454b962d0acf53a96256646667988d1d57d09d")
	require.NoError(t, err)
	input, err := hexutil.Decode("0x2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824")
	require.NoError(t, err)
	hash := GethHash(input)
	require.True(t, bytes.Equal(expected, hash))
}