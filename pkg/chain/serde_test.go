package chain

import (
	"testing"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
	"fmt"
	"encoding/json"
	"github.com/stretchr/testify/suite"
)

type testSigStruct struct {
	Sig Signature
}

var sigStr = "0xc0edd44a50e0110d547064122c5859da6fc6b4339d5d0845f86ff971b64c6a2327e22bf0969c1cde4d0b61acde78da516af5b88cc17ea30e043abb086ccd8c451b"

type signatureSuite struct {
	suite.Suite
	sig Signature
}

func (s *signatureSuite) SetupTest() {
	sigData, err := hexutil.Decode(sigStr)
	require.NoError(s.T(), err)
	var sig Signature
	copy(sig[:], sigData)
	s.sig = sig
}

func (s *signatureSuite) TestSignature_MarshalJSON() {
	data, err := json.Marshal(s.sig)
	require.NoError(s.T(), err)
	require.Equal(s.T(), fmt.Sprintf("\"%s\"", sigStr), string(data))
}

func (s *signatureSuite) TestSignature_UnmarshalJSON() {
	var testSig Signature
	jsonStr := fmt.Sprintf("\"%s\"", sigStr)
	json.Unmarshal([]byte(jsonStr), &testSig)
	require.Equal(s.T(), s.sig, testSig)
}

func (s *signatureSuite) TestSignature_MarshallingInStruct() {
	testStruct := testSigStruct{
		Sig: s.sig,
	}
	data, err := json.Marshal(testStruct)
	require.NoError(s.T(), err)
	var unmarshalledTestStruct testSigStruct
	err = json.Unmarshal(data, &unmarshalledTestStruct)
	require.Equal(s.T(), testStruct.Sig, unmarshalledTestStruct.Sig)
}

func TestSignatureSuite(t *testing.T) {
	suite.Run(t, new(signatureSuite))
}
