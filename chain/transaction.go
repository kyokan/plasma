package chain

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/sha3"
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
	BlkNum  uint64   `json:"BlkNum"`
	TxIdx   uint32   `json:"TxIdx"`
}

type rlpHelper struct {
	BlkNum0   uint64
	TxIdx0    uint32
	OutIdx0   uint8
	Sig0      []byte
	BlkNum1   uint64
	TxIdx1    uint32
	OutIdx1   uint8
	Sig1      []byte
	NewOwner0 common.Address
	Amount0   big.Int
	NewOwner1 common.Address
	Amount1   big.Int
	Fee       big.Int
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

func (tx *Transaction) OutputAt(idx uint8) *Output {
	if idx != 0 && idx != 1 {
		panic(fmt.Sprint("Invalid output index: ", idx))
	}

	if idx == 0 {
		return tx.Output0
	}

	return tx.Output1
}

func (tx *Transaction) OutputFor(addr *common.Address) *Output {
	output := tx.OutputAt(0)

	if util.AddressesEqual(&output.NewOwner, addr) {
		return output
	}

	output = tx.OutputAt(1)

	if util.AddressesEqual(&output.NewOwner, addr) {
		return output
	}

	panic(fmt.Sprint("No output found for address: ", addr.Hex()))
}

func (tx *Transaction) OutputIndexFor(addr *common.Address) uint8 {
	output := tx.OutputAt(0)

	if util.AddressesEqual(&output.NewOwner, addr) {
		return 0
	}

	output = tx.OutputAt(1)

	if util.AddressesEqual(&output.NewOwner, addr) {
		return 1
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

	digest := sha3.Sum256(buf.Bytes())
	return digest[:]
}

func (tx *Transaction) RLPHash() util.Hash {
	bytes, err := rlp.EncodeToBytes(tx)

	if err != nil {
		panic(err)
	}

	return util.DoHash(bytes)
}

func (tx *Transaction) EncodeRLP(w io.Writer) error {
	var itf rlpHelper
	if tx.Input0 != nil {
		itf.BlkNum0 = tx.Input0.BlkNum
		itf.TxIdx0  = tx.Input0.TxIdx
		itf.OutIdx0 = tx.Input0.OutIdx
		itf.Sig0    = tx.Sig0
	}
	if tx.Input1 != nil {
		itf.BlkNum1 = tx.Input1.BlkNum
		itf.TxIdx1  = tx.Input1.TxIdx
		itf.OutIdx1 = tx.Input1.OutIdx
		itf.Sig1    = tx.Sig1
	}
	if tx.Output0 != nil {
		itf.NewOwner0 = tx.Output0.NewOwner
		itf.Amount0   = *tx.Output0.Amount
	}
	if tx.Output1 != nil {
		itf.NewOwner1 = tx.Output1.NewOwner
		itf.Amount1   = *tx.Output1.Amount
	}
	if tx.Fee != nil {
		itf.Fee = *tx.Fee
	}
	return rlp.Encode(w, &itf)
}

func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	var itf rlpHelper
	err := s.Decode(&itf)
	if err != nil {
		return err
	}
	tx.Input0  = NewInput(itf.BlkNum0, itf.TxIdx0, itf.OutIdx0)
	tx.Input1  = NewInput(itf.BlkNum1, itf.TxIdx1, itf.OutIdx1)
	tx.Output0 = NewOutput(itf.NewOwner0, &itf.Amount0)
	tx.Output1 = NewOutput(itf.NewOwner1, &itf.Amount1)
	tx.Sig0 = itf.Sig0
	tx.Sig1 = itf.Sig1
	tx.Fee  = big.NewInt(itf.Fee.Int64())
	return nil
}