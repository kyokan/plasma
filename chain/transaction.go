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
	"log"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Transaction struct {
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

type rlpTransaction struct {
	BlkNum0       *UInt256
	TxIdx0        *UInt256
	OutIdx0       *UInt256
	DepositNonce0 *UInt256
	Sig0          Signature
	BlkNum1       *UInt256
	TxIdx1        *UInt256
	OutIdx1       *UInt256
	DepositNonce1 *UInt256
	Sig1          Signature
	Owner0        common.Address
	Amount0       *UInt256
	Owner1        common.Address
	Amount1       *UInt256
	Fee           *UInt256
}

func ZeroTransaction() *Transaction {
	return &Transaction{
		Input0:  ZeroInput(),
		Input1:  ZeroInput(),
		Output0: ZeroOutput(),
		Output1: ZeroOutput(),
		Fee:     Zero(),
	}
}

func (tx *Transaction) IsDeposit() bool {
	return tx.Output0.IsDeposit()
}

func (tx *Transaction) IsExit() bool {
	return tx != nil &&
			tx.Input1.IsZeroInput() &&
			tx.Output1.IsZeroOutput() &&
			tx.Output0.IsExit()
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

	if util.AddressesEqual(&output.Owner, addr) {
		return output
	}

	output = tx.OutputAt(big.NewInt(1))

	if util.AddressesEqual(&output.Owner, addr) {
		return output
	}

	panic(fmt.Sprint("No output found for address: ", addr.Hex()))
}

func (tx *Transaction) OutputIndexFor(addr *common.Address) *big.Int {
	output := tx.OutputAt(big.NewInt(0))

	if util.AddressesEqual(&output.Owner, addr) {
		return big.NewInt(0)
	}

	output = tx.OutputAt(big.NewInt(1))

	if util.AddressesEqual(&output.Owner, addr) {
		return big.NewInt(1)
	}

	panic(fmt.Sprint("No output found for address: ", addr.Hex()))
}

func (tx *Transaction) Hash(hasher util.Hasher) util.Hash {
	values := []interface{}{
		tx.Input0.Hash(),
		tx.Sig0[:],
		tx.Input1.Hash(),
		tx.Sig1[:],
		tx.Output0.Hash(),
		tx.Output1.Hash(),
		tx.Fee,
		tx.BlkNum,
		tx.TxIdx,
	}

	return doHash(values, hasher)
}

func (tx *Transaction) signatureArray() rlpTransaction {
	return rlpTransaction{
		BlkNum0:       NewUint256(tx.Input0.BlkNum),
		TxIdx0:        NewUint256(tx.Input0.TxIdx),
		OutIdx0:       NewUint256(tx.Input0.OutIdx),
		DepositNonce0: NewUint256(tx.Input0.DepositNonce),
		Sig0:          tx.Sig0,
		BlkNum1:       NewUint256(tx.Input1.BlkNum),
		TxIdx1:        NewUint256(tx.Input1.TxIdx),
		OutIdx1:       NewUint256(tx.Input1.OutIdx),
		DepositNonce1: NewUint256(tx.Input1.DepositNonce),
		Sig1:          tx.Sig1,
		Owner0:        tx.Output0.Owner,
		Amount0:       NewUint256(tx.Output0.Denom),
		Owner1:        tx.Output1.Owner,
		Amount1:       NewUint256(tx.Output1.Denom),
		Fee:           NewUint256(tx.Fee),
	}
}

func (tx *Transaction) SignatureHash() util.Hash {
	encoded, _ := rlp.EncodeToBytes(tx.signatureArray())
	log.Println("encoded", hexutil.Encode(encoded))
	return util.DoHash(encoded)
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
			err = errors.New(fmt.Sprint("invalid component type %v", t))
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
