package validation

import (
	"github.com/stretchr/testify/suite"
	"github.com/kyokan/plasma/pkg/db"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
	"github.com/stretchr/testify/require"
	"testing"
	"github.com/kyokan/plasma/pkg/chain"
	"math/big"
	"github.com/kyokan/plasma/test_util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/pkg/eth"
		)

var depositingAddr = common.HexToAddress("0x627306090abaB3A6e1400e9345bC60c78a8BEf57")

type depositValidationSuite struct {
	suite.Suite
	storage    db.Storage
	ldb        *leveldb.DB
	bwm1       *chain.BlockWithMeta
	mockClient *test_util.EthClientMock
}

func (d *depositValidationSuite) SetupSuite() {
	tmpDir, err := ioutil.TempDir("", "plasma-test")
	require.NoError(d.T(), err)
	ldb, storage, err := db.CreateLevelStorage(tmpDir)
	require.NoError(d.T(), err)
	d.ldb = ldb
	d.storage = storage
}

func (d *depositValidationSuite) TeardownSuite() {
	err := d.ldb.Close()
	require.NoError(d.T(), err)
}

func (d *depositValidationSuite) SetupTest() {
	bwm1, err := inflateFixture("block_1.json")
	require.NoError(d.T(), err)
	d.bwm1 = bwm1
	d.mockClient = &test_util.EthClientMock{}
}

func (d *depositValidationSuite) TestNegativeOutputs() {
	tx := d.bwm1.ConfirmedTransactions[0]
	tx.Transaction.Body.Output0.Amount = big.NewInt(-1)

	err := ValidateDepositTransaction(d.storage, d.mockClient, tx.Transaction)
	require.Error(d.T(), err)
	require.IsType(d.T(), &ErrNegativeOutput{}, err)
	require.Equal(d.T(), uint8(0), err.(*ErrNegativeOutput).Index)

	tx.Transaction.Body.Output0.Amount = big.NewInt(100)
	tx.Transaction.Body.Output1.Amount = big.NewInt(-1)
	err = ValidateSpendTransaction(d.storage, tx.Transaction)
	require.Error(d.T(), err)
	require.IsType(d.T(), &ErrNegativeOutput{}, err)
	require.Equal(d.T(), uint8(1), err.(*ErrNegativeOutput).Index)
}

func (d *depositValidationSuite) TestDefinedInput1() {
	tx := d.bwm1.ConfirmedTransactions[0]
	tx.Transaction.Body.Input1.BlockNumber = 2

	err := ValidateDepositTransaction(d.storage, d.mockClient, tx.Transaction)
	require.Error(d.T(), err)
	require.IsType(d.T(), &ErrDepositDefinedInput1{}, err)

	tx.Transaction.Body.Input1.BlockNumber = 0
	sig, err := randSig()
	require.NoError(d.T(), err)
	tx.Transaction.Body.Input1ConfirmSig = sig

	err = ValidateDepositTransaction(d.storage, d.mockClient, tx.Transaction)
	require.Error(d.T(), err)
	require.IsType(d.T(), &ErrDepositDefinedInput1{}, err)
}

func (d *depositValidationSuite) TestNonEmptyConfirmSig() {
	sig, err := randSig()
	require.NoError(d.T(), err)
	tx := d.bwm1.ConfirmedTransactions[0]
	tx.Transaction.Body.Input0ConfirmSig = sig

	err = ValidateDepositTransaction(d.storage, d.mockClient, tx.Transaction)
	require.Error(d.T(), err)
	require.IsType(d.T(), &ErrDepositNonEmptyConfirmSig{}, err)
}

func (d *depositValidationSuite) TestNonExistentDeposit() {
	var amount *big.Int
	var addr common.Address
	nonce := big.NewInt(1)
	d.mockClient.On("LookupDeposit", nonce).Return(amount, addr, eth.NewErrDepositNotFound(nonce))

	tx := d.bwm1.ConfirmedTransactions[0]
	err := ValidateDepositTransaction(d.storage, d.mockClient, tx.Transaction)
	require.Error(d.T(), err)
	require.IsType(d.T(), &eth.ErrDepositNotFound{}, err)
}

func (d *depositValidationSuite) TestInputOutputValueMismatch() {
	tx := d.bwm1.ConfirmedTransactions[0].Transaction
	mockSuccessfulDeposit(d.mockClient, tx)

	tx.Body.Output0.Amount = tx.Body.Output0.Amount.Mul(tx.Body.Output0.Amount, big.NewInt(10))
	err := ValidateDepositTransaction(d.storage, d.mockClient, tx)
	require.Error(d.T(), err)
	require.IsType(d.T(), &ErrInputOutputValueMismatch{}, err)
}

func (d *depositValidationSuite) TestInvalidSigs_Input0() {
	tx := d.bwm1.ConfirmedTransactions[0].Transaction
	sig, err := randSig()
	require.NoError(d.T(), err)
	tx.Sigs[0] = sig
	mockSuccessfulDeposit(d.mockClient, tx)
	requireInvalidDepositSignature(d.T(), d.storage, d.mockClient, tx, 0)
}

func (d *depositValidationSuite) TestInvalidSigs_Input1() {
	tx := d.bwm1.ConfirmedTransactions[0].Transaction
	sig, err := randSig()
	require.NoError(d.T(), err)
	tx.Sigs[1] = sig
	mockSuccessfulDeposit(d.mockClient, tx)
	requireInvalidDepositSignature(d.T(), d.storage, d.mockClient, tx, 1)
}

// run this test last since it commits things
// to the database
func (d *depositValidationSuite) Test_Z_DoubleSpend() {
	_, err := insertFixture(d.storage, "block_1.json")
	require.NoError(d.T(), err)
	tx := d.bwm1.ConfirmedTransactions[0].Transaction
	mockSuccessfulDeposit(d.mockClient, tx)
	err = ValidateDepositTransaction(d.storage, d.mockClient, tx)
	require.Error(d.T(), err)
	require.IsType(d.T(), &ErrDoubleSpent{}, err)
}

func (d *depositValidationSuite) TestValid() {
	tx := d.bwm1.ConfirmedTransactions[0].Transaction
	mockSuccessfulDeposit(d.mockClient, tx)
	err := ValidateDepositTransaction(d.storage, d.mockClient, tx)
	require.NoError(d.T(), err)
}

func TestValidateDeposit(t *testing.T) {
	suite.Run(t, new(depositValidationSuite))
}

func requireInvalidDepositSignature(t *testing.T, storage db.Storage, client eth.Client, tx *chain.Transaction, inputIndex uint8) {
	err := ValidateDepositTransaction(storage, client, tx)
	require.Error(t, err)
	require.IsType(t, &ErrInvalidSignature{}, err)
	require.Equal(t, inputIndex, err.(*ErrInvalidSignature).InputIndex)
}

func mockSuccessfulDeposit(mock *test_util.EthClientMock, tx *chain.Transaction) {
	// make copy because multiplication below mutates
	sum := big.NewInt(0)
	sum = sum.Add(sum, tx.Body.Output0.Amount).Add(sum, tx.Body.Output1.Amount)
	nonce := big.NewInt(1)
	mock.On("LookupDeposit", nonce).Return(sum, tx.Body.Output1.Owner, nil)
}
