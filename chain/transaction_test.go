package chain

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"
)

func Test_TransactionFullRLP(t *testing.T) {
	tx := Transaction{
		Input0:  randomInput(),
		Input1:  randomInput(),
		Sig0:    randomSig(),
		Sig1:    randomSig(),
		Output0: randomOutput(),
		Output1: randomOutput(),
		Fee:     big.NewInt(rand.Int63()),
		BlkNum:  0, // Not encoded in RLP
		TxIdx:   0, // Not encoded in RLP
	}
	encodeAndDecode(t, tx)
}

func Test_TransactionFirstInputRLP(t *testing.T) {
	tx := Transaction{
		Input0:  randomInput(),
		Input1:  ZeroInput(),
		Sig0:    randomSig(),
		Sig1:    []byte{},
		Output0: randomOutput(),
		Output1: ZeroOutput(),
		Fee:     big.NewInt(rand.Int63()),
		BlkNum:  0,
		TxIdx:   0,
	}
	encodeAndDecode(t, tx)
}

//Helpers
func encodeAndDecode(t *testing.T, tx Transaction) {
	bytes, err := rlp.EncodeToBytes(&tx)
	require.NoError(t, err)
	txDecoded := Transaction{}
	err = rlp.DecodeBytes(bytes, &txDecoded)
	require.NoError(t, err)
	require.Equal(t, tx, txDecoded)
}

func randomInput() *Input {
	return &Input{
		BlkNum: rand.Uint64(),
		TxIdx:  rand.Uint32(),
		OutIdx: uint8(rand.Uint32()),
	}
}

func randomSig() []byte {
	size := 32
	result := make([]byte, size)
	rand.Read(result)
	return result
}

func randomOutput() *Output {
	result := Output{}
	result.Amount = big.NewInt(rand.Int63())
	buf := make([]byte, 20)
	rand.Read(buf)
	for i, _ := range result.NewOwner {
		result.NewOwner[i] = buf[i]
	}
	return &result
}
