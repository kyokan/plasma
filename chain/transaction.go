package chain

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/util"
)

type Deposit struct {
	DepositNonce *big.Int
	Amount       *big.Int
}

type Transaction struct {
	Deposit
	Input0  *Input
	Sig0    Signature
	Input1  *Input
	Sig1    Signature
	Output0 *Output
	Output1 *Output
	Fee     *big.Int
	BlkNum  *big.Int
	TxIdx   *big.Int
}


func ZeroTransaction() *Transaction {
	return &Transaction{
		Deposit: Deposit{ DepositNonce: Zero(), Amount: Zero()},
		Input0: ZeroInput(),
		Input1: ZeroInput(),
		Output0: ZeroOutput(),
		Output1: ZeroOutput(),
		Fee: Zero(),
	}
}

func (tx *Transaction) IsDeposit() bool {
	return tx.DepositNonce != nil && tx.Amount != nil &&
		   tx.DepositNonce.Cmp(Zero()) == 1 && // both greater than zero
	       tx.Amount.Cmp(Zero()) == 1
}

func (tx *Transaction) GetFee() *big.Int {
	return tx.Fee
}

func (tx *Transaction) IsZeroTransaction() bool {
	if tx.IsDeposit() {
		return false
	}
	return tx.Input0.IsZeroInput() &&
		tx.Input1.IsZeroInput() &&
		tx.Output0.IsZeroOutput() &&
		tx.Output1.IsZeroOutput()
}

func (tx *Transaction) InputAt(idx uint8) *Input {
	if idx != 0 && idx != 1 {
		panic(fmt.Sprint("Invalid input index: ", idx))
	}

	if idx == 0 {
		return tx.Input0
	}

	return tx.Input1
}

func (tx *Transaction) OutputAt(idx *big.Int) *Output {
	if idx.Cmp(big.NewInt(0)) == 0 {
		return tx.Output0
	}

	return tx.Output1
}

func (tx *Transaction) OutputFor(addr *common.Address) *Output {
	output := tx.OutputAt(big.NewInt(0))

	if util.AddressesEqual(&output.NewOwner, addr) {
		return output
	}

	output = tx.OutputAt(big.NewInt(1))

	if util.AddressesEqual(&output.NewOwner, addr) {
		return output
	}

	panic(fmt.Sprint("No output found for address: ", addr.Hex()))
}

func (tx *Transaction) OutputIndexFor(addr *common.Address) *big.Int {
	output := tx.OutputAt(big.NewInt(0))

	if util.AddressesEqual(&output.NewOwner, addr) {
		return big.NewInt(0)
	}

	output = tx.OutputAt(big.NewInt(1))

	if util.AddressesEqual(&output.NewOwner, addr) {
		return big.NewInt(1)
	}

	panic(fmt.Sprint("No output found for address: ", addr.Hex()))
}

func (tx *Transaction) Hash(hasher util.Hasher) util.Hash {
	values := []interface{}{
		tx.Input0.Hash(),
		tx.Sig0,
		tx.Input1.Hash(),
		tx.Sig1,
		tx.Output0.Hash(),
		tx.Output1.Hash(),
		tx.Fee,
		tx.BlkNum,
		tx.TxIdx,
	}

	return doHash(values, hasher)
}

func (tx *Transaction) SignatureHash() util.Hash {
	values := []interface{}{
		tx.Input0.Hash(),
		tx.Input1.Hash(),
		tx.Output0.Hash(),
		tx.Output1.Hash(),
		tx.Fee,
	}

	return doHash(values, util.DoHash)
}

func doHash(values []interface{}, hasher util.Hasher) util.Hash {
	buf := new(bytes.Buffer)

	for _, component := range values {
		var err error
		switch t := component.(type) {
		case util.Hash:
			_, err = buf.Write(t)
		case []byte:
			_, err = buf.Write(t)
		case *big.Int:
			_, err = buf.Write(t.Bytes())
		case uint64, uint32:
			err = binary.Write(buf, binary.BigEndian, t)
		default:
			err = errors.New("invalid component type")
		}

		if err != nil {
			panic(err)
		}
	}
	return hasher(buf.Bytes())
}

func (tx *Transaction) RLPHash(hasher util.Hasher) util.Hash {
	bytes, err := rlp.EncodeToBytes(tx)

	if err != nil {
		panic(err)
	}

	return hasher(bytes)
}

func (tx *Transaction) SetIndex(index uint32) {
	tx.TxIdx = big.NewInt(int64(index))
}

