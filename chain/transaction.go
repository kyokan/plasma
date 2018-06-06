package chain

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/util"
)

type Transaction struct {
	Input0  *Input   `rlp:"nil"`
	Input1  *Input   `rlp:"nil"`
	Sig0    []byte
	Sig1    []byte
	Output0 *Output  `rlp:"nil"`
	Output1 *Output  `rlp:"nil"`
	Fee     *big.Int `rlp:"nil"`
	BlkNum  uint64   `rlp:"-"`
	TxIdx   uint32
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