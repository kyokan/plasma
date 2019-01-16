package chain

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"math/rand"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func Test_AddressRLP(t *testing.T) {
	address := RandomAddress()
	encodeAndDecode(t, &address)
}

func TestEmptyAddressRLP(t *testing.T) {
	var address common.Address
	encodeAndDecode(t, &address)
}

func Test_SignatureRLP(t *testing.T) {
	signature := RandomConfirmationSig()
	encodeAndDecode(t, &signature)
}

func Test_EmptySignatureRLP(t *testing.T) {
	var signature Signature
	encodeAndDecode(t, &signature)
}

func Test_BigIntRLP(t *testing.T) {
	input := big.NewInt(rand.Int63())
	value := NewUint256(input)
	encodeAndDecode(t, &value)
	output := value.ToBig()
	require.Equal(t, input, output)
}

func Test_ZeroBigIntRLP(t *testing.T) {
	input := big.NewInt(0)
	value := NewUint256(input)
	encodeAndDecode(t, &value)
	output := value.ToBig()
	require.Equal(t, input, output)
}

func Test_NilBigIntRLP(t *testing.T) {
	zero := big.NewInt(0)
	var input *big.Int
	value := NewUint256(input)
	encodeAndDecode(t, &value)
	output := value.ToBig()
	require.Equal(t, zero, output)
}

func Test_TransactionFullRLP(t *testing.T) {
	tx := Transaction{
		Input0:  RandomInput(),
		Input1:  RandomInput(),
		Sig0:    RandomConfirmationSig(),
		Sig1:    RandomConfirmationSig(),
		Output0: RandomOutput(),
		Output1: RandomOutput(),
		Fee:     big.NewInt(rand.Int63()),
		BlkNum:  nil, // Not encoded in RLP
		TxIdx:   nil, // Not encoded in RLP
	}
	encodeAndDecode(t, &tx)
}

func Test_TransactionFirstInputRLP(t *testing.T) {
	tx := Transaction{
		Input0:  RandomInput(),
		Input1:  ZeroInput(),
		Sig0:    RandomConfirmationSig(),
		Output0: RandomOutput(),
		Output1: ZeroOutput(),
		Fee:     big.NewInt(rand.Int63()),
		BlkNum:  nil,
		TxIdx:   nil,
	}
	encodeAndDecode(t, &tx)
}

func Test_ConfirmedTransactionRLP(t *testing.T) {
	confirmed := ConfirmedTransaction{
		Transaction: Transaction{
			Input0:  RandomInput(),
			Input1:  RandomInput(),
			Sig0:    RandomConfirmationSig(),
			Sig1:    RandomConfirmationSig(),
			Output0: RandomOutput(),
			Output1: RandomOutput(),
			Fee:     big.NewInt(rand.Int63()),
			BlkNum:  nil, // Not encoded in RLP
			TxIdx:   nil, // Not encoded in RLP
		},
		Signatures: [2]Signature{RandomConfirmationSig(), RandomConfirmationSig()},
	}
	encodeAndDecode(t, &confirmed)
}

func Test_ConfirmedTransactionFirstInputRLP(t *testing.T) {
	confirmed := ConfirmedTransaction{
		Transaction: Transaction{
			Input0:  RandomInput(),
			Input1:  ZeroInput(),
			Sig0:    RandomConfirmationSig(),
			Sig1:    RandomConfirmationSig(),
			Output0: RandomOutput(),
			Output1: RandomOutput(),
			Fee:     big.NewInt(rand.Int63()),
			BlkNum:  nil, // Not encoded in RLP
			TxIdx:   nil, // Not encoded in RLP
		},
		Signatures: [2]Signature{RandomConfirmationSig(), },
	}
	encodeAndDecode(t, &confirmed)
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
