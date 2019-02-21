package validation

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"github.com/kyokan/plasma/pkg/db"
	"io/ioutil"
	"github.com/kyokan/plasma/pkg/chain"
	"encoding/json"
	"math/big"
	"github.com/syndtr/goleveldb/leveldb"
	"fmt"
	"github.com/stretchr/testify/require"
	"crypto/rand"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/kyokan/plasma/pkg/eth"
)

type spendValidationSuite struct {
	suite.Suite
	storage db.Storage
	ldb     *leveldb.DB
	bwm1    *chain.BlockWithMeta
	bwm2    *chain.BlockWithMeta
	key     *ecdsa.PrivateKey
}

func (v *spendValidationSuite) SetupSuite() {
	tmpDir, err := ioutil.TempDir("", "plasma-test")
	require.NoError(v.T(), err)
	ldb, storage, err := db.CreateLevelStorage(tmpDir)
	require.NoError(v.T(), err)
	v.ldb = ldb
	v.storage = storage

	_, err = insertFixture(v.storage, "block_1.json")
	require.NoError(v.T(), err)
	_, err = insertFixture(v.storage, "block_2.json")
	require.NoError(v.T(), err)

	key, err := crypto.HexToECDSA("c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3")
	require.NoError(v.T(), err)
	v.key = key
}

func (v *spendValidationSuite) SetupTest() {
	bwm1, err := inflateFixture("block_2.json")
	require.NoError(v.T(), err)
	v.bwm1 = bwm1
	bwm2, err := inflateFixture("block_3.json")
	require.NoError(v.T(), err)
	v.bwm2 = bwm2
}

func (v *spendValidationSuite) TeardownSuite() {
	err := v.ldb.Close()
	require.NoError(v.T(), err)
}

func (v *spendValidationSuite) TestNegativeOutputs() {
	tx := v.bwm1.ConfirmedTransactions[0]
	tx.Transaction.Body.Output0.Amount = big.NewInt(-1)

	err := ValidateSpendTransaction(v.storage, tx.Transaction)
	require.Error(v.T(), err)
	require.IsType(v.T(), &ErrNegativeOutput{}, err)
	require.Equal(v.T(), uint8(0), err.(*ErrNegativeOutput).Index)

	tx.Transaction.Body.Output0.Amount = big.NewInt(100)
	tx.Transaction.Body.Output1.Amount = big.NewInt(-1)
	err = ValidateSpendTransaction(v.storage, tx.Transaction)
	require.Error(v.T(), err)
	require.IsType(v.T(), &ErrNegativeOutput{}, err)
	require.Equal(v.T(), uint8(1), err.(*ErrNegativeOutput).Index)
}

func (v *spendValidationSuite) TestTxNotFound_Input0() {
	tx := v.bwm1.ConfirmedTransactions[0].Transaction
	tx.Body.Input0.BlockNumber = 10
	requireNotFound(v.T(), v.storage, tx, 0)
}

func (v *spendValidationSuite) TestTxNotFound_Input1() {
	tx := v.bwm1.ConfirmedTransactions[0].Transaction
	tx.Body.Input1.BlockNumber = 10
	err := reSign(tx, v.key, 0)
	require.NoError(v.T(), err)
	requireNotFound(v.T(), v.storage, tx, 1)
}

func (v *spendValidationSuite) TestMismatchedConfirmSigs_Input0() {
	var sig chain.Signature
	fillRand(sig[:])
	tx := v.bwm1.ConfirmedTransactions[0].Transaction
	tx.Body.Input0ConfirmSig = sig
	requireMismatchedConfirmSigs(v.T(), v.storage, tx, 0)
}

func (v *spendValidationSuite) TestMismatchedConfirmSigs_Input1() {
	var sig chain.Signature
	fillRand(sig[:])
	tx := v.bwm1.ConfirmedTransactions[0].Transaction
	// need to re-sign here to get past the signature checker
	// on input 0
	tx.Body.Input1.BlockNumber = 1
	tx.Body.Input1.TransactionIndex = 0
	tx.Body.Input1ConfirmSig = sig
	err := reSign(tx, v.key, 0)
	require.NoError(v.T(), err)
	requireMismatchedConfirmSigs(v.T(), v.storage, tx, 1)
}

func (v *spendValidationSuite) TestInvalidSigs_Input0() {
	var sig chain.Signature
	fillRand(sig[:])
	tx := v.bwm1.ConfirmedTransactions[0].Transaction
	tx.Sigs[0] = sig
	requireInvalidSignature(v.T(), v.storage, tx, 0)
}

func (v *spendValidationSuite) TestInvalidSigs_Input1() {
	tx := v.bwm1.ConfirmedTransactions[0].Transaction
	// need to make input1 non-zero. this will break the signature, which
	// is what we want. block 1 corresponds to input 0.
	tx.Body.Input1.BlockNumber = 1
	tx.Body.Input1.TransactionIndex = 0
	tx.Body.Input1ConfirmSig = tx.Body.Input0ConfirmSig
	err := reSign(tx, v.key, 0)
	require.NoError(v.T(), err)
	requireInvalidSignature(v.T(), v.storage, tx, 1)
}

func (v *spendValidationSuite) TestInputOutputValueMismatch() {
	tx := v.bwm1.ConfirmedTransactions[0].Transaction
	tx.Body.Output0.Amount = tx.Body.Output0.Amount.Mul(tx.Body.Output0.Amount, big.NewInt(10))
	err := reSign(tx, v.key, 0)
	require.NoError(v.T(), err)
	err = reSign(tx, v.key, 1)
	require.NoError(v.T(), err)
	err = ValidateSpendTransaction(v.storage, tx)
	require.Error(v.T(), err)
	require.IsType(v.T(), &ErrInputOutputValueMismatch{}, err)
}

func (v *spendValidationSuite) TestDoubleSpend() {
	tx := v.bwm1.ConfirmedTransactions[0].Transaction
	err := ValidateSpendTransaction(v.storage, tx)
	require.Error(v.T(), err)
	require.IsType(v.T(), &ErrDoubleSpent{}, err)
}

func (v *spendValidationSuite) TestIdenticalInputs() {
	tx := v.bwm1.ConfirmedTransactions[0].Transaction
	tx.Body.Input1 = tx.Body.Input0
	err := reSign(tx, v.key, 0)
	require.NoError(v.T(), err)
	err = reSign(tx, v.key, 1)
	require.NoError(v.T(), err)
	err = ValidateSpendTransaction(v.storage, tx)
	require.Error(v.T(), err)
	require.IsType(v.T(), &ErrIdenticalInputs{}, err)
}

func (v *spendValidationSuite) TestValid() {
	tx := v.bwm2.ConfirmedTransactions[0].Transaction
	err := ValidateSpendTransaction(v.storage, tx)
	require.NoError(v.T(), err)
}

func TestValidateSpend(t *testing.T) {
	suite.Run(t, new(spendValidationSuite))
}

func requireNotFound(t *testing.T, storage db.Storage, tx *chain.Transaction, inputIndex uint8) {
	err := ValidateSpendTransaction(storage, tx)
	require.Error(t, err)
	require.IsType(t, &ErrTxNotFound{}, err)
	input := tx.Body.InputAt(inputIndex)
	require.Equal(t, inputIndex, err.(*ErrTxNotFound).InputIndex)
	require.Equal(t, input.BlockNumber, err.(*ErrTxNotFound).BlockNumber)
	require.Equal(t, input.TransactionIndex, err.(*ErrTxNotFound).TransactionIndex)
}

func requireMismatchedConfirmSigs(t *testing.T, storage db.Storage, tx *chain.Transaction, inputIndex uint8) {
	err := ValidateSpendTransaction(storage, tx)
	require.Error(t, err)
	require.IsType(t, &ErrConfirmSigMismatch{}, err)
	require.Equal(t, inputIndex, err.(*ErrConfirmSigMismatch).InputIndex)
}

func requireInvalidSignature(t *testing.T, storage db.Storage, tx *chain.Transaction, inputIndex uint8) {
	err := ValidateSpendTransaction(storage, tx)
	require.Error(t, err)
	require.IsType(t, &ErrInvalidSignature{}, err)
	require.Equal(t, inputIndex, err.(*ErrInvalidSignature).InputIndex)
}

func insertFixture(storage db.Storage, fixtureName string) (*chain.BlockWithMeta, error) {
	bwm, err := inflateFixture(fixtureName)
	if err != nil {
		return nil, err
	}
	err = storage.InsertBlock(bwm.Block, bwm.Metadata, bwm.ConfirmedTransactions)
	if err != nil {
		return nil, err
	}
	return bwm, nil
}

func inflateFixture(fixtureName string) (*chain.BlockWithMeta, error) {
	fixture, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s", fixtureName))
	if err != nil {
		return nil, err
	}
	return unmarshalBlockWithMeta(fixture)
}

func fillRand(buf []byte) {
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
}

func unmarshalBlockWithMeta(fixture []byte) (*chain.BlockWithMeta, error) {
	var bwm chain.BlockWithMeta
	err := json.Unmarshal(fixture, &bwm)
	if err != nil {
		return nil, err
	}
	return &bwm, nil
}

func reSign(tx *chain.Transaction, key *ecdsa.PrivateKey, index int) error {
	hash := tx.Body.SignatureHash()
	sig, err := eth.Sign(key, hash)
	if err != nil {
		return err
	}
	tx.Sigs[index] = sig
	return nil
}
