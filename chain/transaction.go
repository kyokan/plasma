package chain

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/util"
)

// JSON tags needed for test fixtures
type Transaction struct {
	Input0  *Input   `json:"Input0"`
	Sig0    []byte   `json:"Sig0"`
	Input1  *Input   `json:"Input1"`
	Sig1    []byte   `json:"Sig1"`
	Output0 *Output  `json:"Output0"`
	Output1 *Output  `json:"Output1"`
	Fee     *big.Int `json:"Fee"`
	BlkNum  *big.Int `json:"BlkNum"`
	TxIdx   *big.Int `json:"TxIdx"`
	RootSig []byte   `json:"RootSig"`
}

// Transaction encoding:
// [Blknum0, TxIndex0, Oindex0, depositNonce0, Amount0, ConfirmSig0
//  Blknum1, TxIndex1, Oindex1, depositNonce1, Amount1, ConfirmSig1
//  NewOwner0, Denom0, NewOwner1, Denom1, Fee]
type rlpHelper struct {
	BlkNum0       big.Int
	TxIdx0        big.Int
	OutIdx0       big.Int
	DepositNonce0 big.Int
	Amount0       big.Int
	Sig0          []byte

	BlkNum1       big.Int
	TxIdx1        big.Int
	OutIdx1       big.Int
	DepositNonce1 big.Int
	Amount1       big.Int
	Sig1          []byte

	NewOwner0     common.Address
	Denom0        big.Int

	NewOwner1     common.Address
	Denom1        big.Int

	Fee           big.Int
	RootSig       []byte
}

func ZeroTransaction() *Transaction {
	return &Transaction{
		Input0: ZeroInput(),
		Input1: ZeroInput(),
		Output0: ZeroOutput(),
		Output1: ZeroOutput(),
		Fee: big.NewInt(0),
	}
}

func (tx *Transaction) IsDeposit() bool {
	return tx.Input0.IsZeroInput() &&
		tx.Input1.IsZeroInput() &&
		!tx.Output0.IsZeroOutput() &&
		tx.Output1.IsZeroOutput()
}

func (tx *Transaction) IsZeroTransaction() bool {
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

func (tx *Transaction) Hash() util.Hash {
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

	return doHash(values)
}

func (tx *Transaction) SignatureHash() util.Hash {
	values := []interface{}{
		tx.Input0.Hash(),
		tx.Input1.Hash(),
		tx.Output0.Hash(),
		tx.Output1.Hash(),
		tx.Fee,
	}

	return doHash(values)
}

func doHash(values []interface{}) util.Hash {
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
	digest := util.DoHash(buf.Bytes())
	return digest
}

func (tx *Transaction) RLPHash() util.Hash {
	bytes, err := rlp.EncodeToBytes(tx)

	if err != nil {
		panic(err)
	}

	return util.DoHash(bytes)
}

func (tx *Transaction) SetIndex(index uint32) {
	tx.TxIdx = big.NewInt(int64(index))
}

func (tx *Transaction) EncodeRLP(w io.Writer) error {
	var itf rlpHelper
	if tx.Input0 != nil {
		itf.BlkNum0 = *tx.Input0.BlkNum
		itf.TxIdx0  = *tx.Input0.TxIdx
		itf.OutIdx0 = *tx.Input0.OutIdx
		itf.Sig0    = tx.Sig0
	}
	if tx.Input1 != nil {
		itf.BlkNum1 = *tx.Input1.BlkNum
		itf.TxIdx1  = *tx.Input1.TxIdx
		itf.OutIdx1 = *tx.Input1.OutIdx
		itf.Sig1    = tx.Sig1
	}
	if tx.Output0 != nil {
		itf.NewOwner0 = tx.Output0.NewOwner
		itf.Amount0   = *tx.Output0.Denom
	}
	if tx.Output1 != nil {
		itf.NewOwner1 = tx.Output1.NewOwner
		itf.Amount1   = *tx.Output1.Denom
	}
	if tx.Fee != nil {
		itf.Fee = *tx.Fee
	}
	itf.RootSig = tx.RootSig
	return rlp.Encode(w, &itf)
}

func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	var itf rlpHelper
	err := s.Decode(&itf)
	if err != nil {
		return err
	}
	tx.Input0  = NewInput(&itf.BlkNum0, &itf.TxIdx0, &itf.OutIdx0)
	tx.Input1  = NewInput(&itf.BlkNum1, &itf.TxIdx1, &itf.OutIdx1)
	tx.Output0 = NewOutput(itf.NewOwner0, &itf.Amount0)
	tx.Output0.DepositNonce = &itf.DepositNonce0
	tx.Output1 = NewOutput(itf.NewOwner1, &itf.Amount1)
	tx.Output1.DepositNonce = &itf.DepositNonce1
	tx.Sig0 = itf.Sig0
	tx.Sig1 = itf.Sig1
	tx.Fee  = big.NewInt(itf.Fee.Int64())
	tx.RootSig = itf.RootSig
	return nil
}