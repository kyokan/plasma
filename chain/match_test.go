package chain

import (
    "math/big"
    "math/rand"
    "testing"
    "time"

    "github.com/kyokan/plasma/common/mocks"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

const size = 10000
const max  = 4 * size

func Test_OneTransactionMatches(t *testing.T) {
    rand.Seed(time.Now().Unix())
    from   := randomAddress()
    to     := randomAddress()
    amount := big.NewInt(0)
    client := new(mocks.Client)
    client.On("SignData", mock.AnythingOfType("*common.Address"), mock.AnythingOfType("[]uint8")).Return(randomSig(), nil)
    transactions := make([]Transaction, size)
    idx := rand.Intn(size)
    for i := 0; i < size; i++ {
        outputIdx := rand.Float32() < 0.5
        output := &Output{
            NewOwner: from,
            Amount: big.NewInt(int64(rand.Intn(max))),
        }
        if i == idx {
            amount.Add(amount, output.Amount)
        }
        transactions[i] = Transaction{
            Input0: randomInput(),
            Input1: randomInput(),
        }
        if outputIdx == false {
            transactions[i].Output0 = output
            transactions[i].Output1 = randomOutput()
        } else {
            transactions[i].Output1 = output
            transactions[i].Output0 = randomOutput()
        }
    }
    tx, err := FindBestUTXOs(from, to, amount, transactions, client)
    require.NoError(t, err)
    require.Equal(t, ZeroOutput(), tx.Output1)
    require.Equal(t, 0, amount.Cmp(tx.Output0.Amount))
}

func Test_TwoTransactionsMatch(t *testing.T) {
    rand.Seed(time.Now().Unix())
    from   := randomAddress()
    to     := randomAddress()
    amount := big.NewInt(0)
    client := new(mocks.Client)
    client.On("SignData", mock.AnythingOfType("*common.Address"), mock.AnythingOfType("[]uint8")).Return(randomSig(), nil)
    transactions := make([]Transaction, size)
    firstIdx := rand.Intn(size)
    secondIdx := rand.Intn(size)
    for ; firstIdx == secondIdx; {
        secondIdx = rand.Intn(size)
    }
    for i := 0; i < size; i++ {
        outputIdx := rand.Float32() < 0.5
        output := &Output{
            NewOwner: from,
            Amount: big.NewInt(int64(rand.Intn(max))),
        }
        if i == firstIdx || i == secondIdx {
            amount.Add(amount, output.Amount)
        }
        transactions[i] = Transaction{
            Input0: randomInput(),
            Input1: randomInput(),
        }
        if outputIdx == false {
            transactions[i].Output0 = output
            transactions[i].Output1 = randomOutput()
        } else {
            transactions[i].Output1 = output
            transactions[i].Output0 = randomOutput()
        }
    }
    tx, err := FindBestUTXOs(from, to, amount, transactions, client)
    require.NoError(t, err)
    require.Equal(t, ZeroOutput(), tx.Output1)
    require.Equal(t, 0, amount.Cmp(tx.Output0.Amount))
}

func Test_AmountLessThanMinTransaction(t *testing.T) {
    rand.Seed(time.Now().Unix())
    from   := randomAddress()
    to     := randomAddress()
    amount := big.NewInt(4)
    client := new(mocks.Client)
    client.On("SignData", mock.AnythingOfType("*common.Address"), mock.AnythingOfType("[]uint8")).Return(randomSig(), nil)
    transactions := make([]Transaction, size)
    for i := 0; i < size; i++ {
        outputIdx := rand.Float32() < 0.5
        output := &Output{
            NewOwner: from,
            Amount: big.NewInt(int64(5 + rand.Intn(max))),
        }
        transactions[i] = Transaction{
            Input0: randomInput(),
            Input1: randomInput(),
        }
        if outputIdx == false {
            transactions[i].Output0 = output
            transactions[i].Output1 = randomOutput()
        } else {
            transactions[i].Output1 = output
            transactions[i].Output0 = randomOutput()
        }
    }
    tx, err := FindBestUTXOs(from, to, amount, transactions, client)
    require.NoError(t, err)
    require.Equal(t, from, tx.Output1.NewOwner)
    require.Equal(t, 0, amount.Cmp(tx.Output0.Amount))
}

func Test_AmountLessThanTwoTransactions(t *testing.T) {
    rand.Seed(time.Now().Unix())
    from   := randomAddress()
    to     := randomAddress()
    amount := big.NewInt(0)
    client := new(mocks.Client)
    client.On("SignData", mock.AnythingOfType("*common.Address"), mock.AnythingOfType("[]uint8")).Return(randomSig(), nil)
    transactions := make([]Transaction, size)
    firstIdx := rand.Intn(size)
    secondIdx := rand.Intn(size)
    for ; firstIdx == secondIdx; {
        secondIdx = rand.Intn(size)
    }
    for i := 0; i < size; i++ {
        outputIdx := rand.Float32() < 0.5
        output := &Output{
            NewOwner: from,
            Amount: big.NewInt(int64(10 * rand.Intn(max))),
        }
        if i == firstIdx || i == secondIdx {
            amount.Add(amount, output.Amount)
        }
        transactions[i] = Transaction{
            Input0: randomInput(),
            Input1: randomInput(),
        }
        if outputIdx == false {
            transactions[i].Output0 = output
            transactions[i].Output1 = randomOutput()
        } else {
            transactions[i].Output1 = output
            transactions[i].Output0 = randomOutput()
        }
    }
    amount.Sub(amount, big.NewInt(1))
    tx, err := FindBestUTXOs(from, to, amount, transactions, client)
    require.NoError(t, err)
    require.Equal(t, from, tx.Output1.NewOwner)
    require.Equal(t, 0, amount.Cmp(tx.Output0.Amount))
}

func Test_NoMatch(t *testing.T) {
    rand.Seed(time.Now().Unix())
    from   := randomAddress()
    to     := randomAddress()
    amount := big.NewInt(1 + 2 * max)
    client := new(mocks.Client)
    client.On("SignData", mock.AnythingOfType("*common.Address"), mock.AnythingOfType("[]uint8")).Return(randomSig(), nil)
    transactions := make([]Transaction, size)
    for i := 0; i < size; i++ {
        outputIdx := rand.Float32() < 0.5
        output := &Output{
            NewOwner: from,
            Amount: big.NewInt(int64(rand.Intn(max))),
        }
        transactions[i] = Transaction{
            Input0: randomInput(),
            Input1: randomInput(),
        }
        if outputIdx == false {
            transactions[i].Output0 = output
            transactions[i].Output1 = randomOutput()
        } else {
            transactions[i].Output1 = output
            transactions[i].Output0 = randomOutput()
        }
    }
    tx, err := FindBestUTXOs(from, to, amount, transactions, client)
    require.Error(t, err)
    require.Nil(t, tx)
}

func Test_NoInput(t *testing.T) {
    rand.Seed(time.Now().Unix())
    from   := randomAddress()
    to     := randomAddress()
    amount := big.NewInt(101)
    client := new(mocks.Client)
    client.On("SignData", mock.AnythingOfType("*common.Address"), mock.AnythingOfType("[]uint8")).Return(randomSig(), nil)
    transactions := make([]Transaction, 0, size)
    tx, err := FindBestUTXOs(from, to, amount, transactions, client)
    require.Error(t, err)
    require.Nil(t, tx)
}

func Test_OneInputLessThanAmount(t *testing.T) {
    rand.Seed(time.Now().Unix())
    from   := randomAddress()
    to     := randomAddress()
    amount := big.NewInt(max + 1)
    client := new(mocks.Client)
    client.On("SignData", mock.AnythingOfType("*common.Address"), mock.AnythingOfType("[]uint8")).Return(randomSig(), nil)
    size := 1
    transactions := make([]Transaction, size)
    for i := 0; i < size; i++ {
        outputIdx := rand.Float32() < 0.5
        output := &Output{
            NewOwner: from,
            Amount: big.NewInt(int64(rand.Intn(max))),
        }
        transactions[i] = Transaction{
            Input0: randomInput(),
            Input1: randomInput(),
        }
        if outputIdx == false {
            transactions[i].Output0 = output
            transactions[i].Output1 = randomOutput()
        } else {
            transactions[i].Output1 = output
            transactions[i].Output0 = randomOutput()
        }
    }
    tx, err := FindBestUTXOs(from, to, amount, transactions, client)
    require.Error(t, err)
    require.Nil(t, tx)
}

func Test_OneInput(t *testing.T) {
    rand.Seed(time.Now().Unix())
    from   := randomAddress()
    to     := randomAddress()
    amount := big.NewInt(0)
    client := new(mocks.Client)
    client.On("SignData", mock.AnythingOfType("*common.Address"), mock.AnythingOfType("[]uint8")).Return(randomSig(), nil)
    transactions := make([]Transaction, 1)
    for i := 0; i < 1; i++ {
        outputIdx := rand.Float32() < 0.5
        output := &Output{
            NewOwner: from,
            Amount: big.NewInt(int64(5 + rand.Intn(max))),
        }
        amount.Sub(output.Amount, big.NewInt(4))
        transactions[i] = Transaction{
            Input0: randomInput(),
            Input1: randomInput(),
        }
        if outputIdx == false {
            transactions[i].Output0 = output
            transactions[i].Output1 = randomOutput()
        } else {
            transactions[i].Output1 = output
            transactions[i].Output0 = randomOutput()
        }
    }
    tx, err := FindBestUTXOs(from, to, amount, transactions, client)
    require.NoError(t, err)
    require.Equal(t, from, tx.Output1.NewOwner)
    require.Equal(t, 0, amount.Cmp(tx.Output0.Amount))
}
