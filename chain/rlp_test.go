package chain

import (
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"
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
	encodeAndDecode(t, &tx)
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
	encodeAndDecode(t, &tx)
}

func Test_InputRLP(t *testing.T) {
	input := randomInput()
	encodeAndDecode(t, &input)
}

func Test_ZeroInputRLP(t *testing.T) {
	input := ZeroOutput()
	encodeAndDecode(t, &input)
}

func Test_OuputRLP(t *testing.T) {
	output := randomOutput()
	encodeAndDecode(t, &output)
}

func Test_ZeroOuputRLP(t *testing.T) {
	output := ZeroOutput()
	encodeAndDecode(t, &output)
}

func Test_BlockRLP(t *testing.T) {
	blk := Block{
		Header: &BlockHeader{
			MerkleRoot: randomSig(),
			RLPMerkleRoot: randomSig(),
			PrevHash: randomSig(),
			Number: rand.Uint64(),
		},
		BlockHash: randomSig(),
	}
	encodeAndDecode(t, &blk)
}

func Test_FlowRLP(t *testing.T) {
	flow := Flow{
		BlkNum: rand.Uint64(),
		TxIdx: rand.Uint32(),
		OutIdx: uint8(rand.Uint32()),
		Hash: randomSig(),
	}
	encodeAndDecode(t, &flow)
}

//Helpers
func encodeAndDecode(t *testing.T, itf interface{}) {
	v := reflect.ValueOf(itf)
	kind := v.Type().Kind()
	if kind != reflect.Ptr {
		require.NoError(t, errors.New(fmt.Sprintf("expecting pointer, got %v", v.Type())))
	}

	bytes, err := rlp.EncodeToBytes(itf)
	require.NoError(t, err)
	reflected := reflect.New(v.Elem().Type())
	itfDecoded := reflected.Interface()
	err = rlp.DecodeBytes(bytes, reflected.Interface())
	require.NoError(t, err)
	require.Equal(t, itf, itfDecoded)
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

func randomAddress() common.Address {
	buf := make([]byte, 20)
	rand.Read(buf)
	return common.BytesToAddress(buf)
}
