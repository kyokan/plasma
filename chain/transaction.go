package chain

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/keybase/go-codec/codec"
	"github.com/kyokan/plasma/util"
	"math/big"
	"github.com/ethereum/go-ethereum/common"
)

type Transaction struct {
	Input0  *Input
	Input1  *Input
	Sig0    []byte
	Sig1    []byte
	Output0 *Output
	Output1 *Output
	Fee     *big.Int
	BlkNum  uint64
	TxIdx   uint32
}

func TransactionFromCbor(data []byte) (*Transaction, error) {
	hdl := util.PatchedCBORHandle()
	dec := codec.NewDecoderBytes(data, hdl)
	ptr := &Transaction{}
	err := dec.Decode(ptr)

	if err != nil {
		return nil, err
	}

	return ptr, nil
}

func (tx *Transaction) IsDeposit() bool {
	return tx.Input0.IsZeroInput() &&
		tx.Input1.IsZeroInput() &&
		!tx.Output0.IsZeroOutput() &&
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

func (tx *Transaction) ToCbor() ([]byte, error) {
	buf := new(bytes.Buffer)
	bw := bufio.NewWriter(buf)
	hdl := util.PatchedCBORHandle()
	enc := codec.NewEncoder(bw, hdl)
	err := enc.Encode(tx)

	if err != nil {
		return nil, err
	}

	bw.Flush()

	return buf.Bytes(), nil
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
