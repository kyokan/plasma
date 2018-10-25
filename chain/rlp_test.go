package chain

import (
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func Test_TransactionFullRLP(t *testing.T) {
	tx := Transaction{
		Input0:  RandomInput(),
		Input1:  RandomInput(),
		Sig0:    RandomSig(),
		Sig1:    RandomSig(),
		Output0: RandomOutput(),
		Output1: RandomOutput(),
		Fee:     big.NewInt(rand.Int63()),
		BlkNum:  0, // Not encoded in RLP
		TxIdx:   0, // Not encoded in RLP
		RootSig: RandomSig(),
	}
	encodeAndDecode(t, &tx)
}

func Test_TransactionFirstInputRLP(t *testing.T) {
	tx := Transaction{
		Input0:  RandomInput(),
		Input1:  ZeroInput(),
		Sig0:    RandomSig(),
		Sig1:    []byte{},
		Output0: RandomOutput(),
		Output1: ZeroOutput(),
		Fee:     big.NewInt(rand.Int63()),
		BlkNum:  0,
		TxIdx:   0,
		RootSig: RandomSig(),
	}
	encodeAndDecode(t, &tx)
}

func Test_InputRLP(t *testing.T) {
	input := RandomInput()
	encodeAndDecode(t, &input)
}

func Test_ZeroInputRLP(t *testing.T) {
	input := ZeroOutput()
	encodeAndDecode(t, &input)
}

func Test_OuputRLP(t *testing.T) {
	output := RandomOutput()
	encodeAndDecode(t, &output)
}

func Test_ZeroOuputRLP(t *testing.T) {
	output := ZeroOutput()
	encodeAndDecode(t, &output)
}

func Test_BlockRLP(t *testing.T) {
	blk := Block{
		Header: &BlockHeader{
			MerkleRoot:    RandomSig(),
			RLPMerkleRoot: RandomSig(),
			PrevHash:      RandomSig(),
			Number:        rand.Uint64(),
		},
		BlockHash: RandomSig(),
	}
	encodeAndDecode(t, &blk)
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
