package chain

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"io"
	"math/big"
	"github.com/kyokan/plasma/util"
)

type Signature [65]byte

type UInt256 [32]byte

func NewUint256(i *big.Int) *UInt256 {
	var result UInt256
	if i != nil {
		bytes := i.Bytes()
		diff := len(result) - len(bytes)
		for i := 0; i != len(bytes); i++ {
			result[diff+i] = bytes[i]
		}
	}
	return &result
}

func (uint *UInt256) ToBig() *big.Int {
	result := big.NewInt(0)
	if uint != nil {
		firstNonZero := 0
		for ; firstNonZero != len(*uint); firstNonZero++ {
			if uint[firstNonZero] != 0 {
				break
			}
		}
		result.SetBytes((*uint)[firstNonZero:])
	}
	return result
}

// Transaction encoding:
// [Blknum1, TxIndex1, Oindex1, DepositNonce1, Owner1, Input1ConfirmSig,
//  Blknum2, TxIndex2, Oindex2, DepositNonce2, Owner2, Input2ConfirmSig,
//  Owner, Denom1, Owner, Denom2, Fee]
type rlpTransactionHelper struct {
	BlkNum0       *UInt256       // input0
	TxIdx0        *UInt256       // input0
	OutIdx0       *UInt256       // input0
	DepositNonce0 *UInt256       // input0
	Owner0        common.Address // input0
	Sig0          Signature      // signature over input0

	BlkNum1       *UInt256       // input1
	TxIdx1        *UInt256       // input1
	OutIdx1       *UInt256       // input1
	DepositNonce1 *UInt256       // input1
	Owner1        common.Address // input1
	Sig1          Signature      // signature over input1

	NewOwner0 common.Address // output0
	Denom0    *UInt256       // output0

	NewOwner1 common.Address // output1
	Denom1    *UInt256       // output1

	Fee *UInt256 // transaction
}

func (tx *Transaction) EncodeRLP(w io.Writer) error {
	var itf rlpTransactionHelper
	if tx.Input0 != nil {
		itf.BlkNum0 = NewUint256(util.Uint642Big(tx.Input0.BlkNum))
		itf.TxIdx0 = NewUint256(util.Uint322Big(tx.Input0.TxIdx))
		itf.OutIdx0 = NewUint256(util.Uint82Big(tx.Input0.OutIdx))
		itf.DepositNonce0 = NewUint256(tx.Input0.DepositNonce)
		itf.Owner0 = tx.Input0.Owner
		itf.Sig0 = tx.Sig0
	} else {
		itf.BlkNum0 = NewUint256(nil)
		itf.TxIdx0 = NewUint256(nil)
		itf.OutIdx0 = NewUint256(nil)
		itf.DepositNonce0 = NewUint256(nil)
	}
	if tx.Input1 != nil {
		itf.BlkNum1 = NewUint256(util.Uint642Big(tx.Input1.BlkNum))
		itf.TxIdx1 = NewUint256(util.Uint322Big(tx.Input1.TxIdx))
		itf.OutIdx1 = NewUint256(util.Uint82Big(tx.Input1.OutIdx))
		itf.DepositNonce1 = NewUint256(tx.Input1.DepositNonce)
		itf.Owner1 = tx.Input1.Owner
		itf.Sig1 = tx.Sig1
	} else {
		itf.BlkNum1 = NewUint256(nil)
		itf.TxIdx1 = NewUint256(nil)
		itf.OutIdx1 = NewUint256(nil)
		itf.DepositNonce1 = NewUint256(nil)
	}
	if tx.Output0 != nil {
		itf.NewOwner0 = tx.Output0.Owner
		itf.Denom0 = NewUint256(tx.Output0.Denom)
	} else {
		itf.Denom0 = NewUint256(nil)
	}
	if tx.Output1 != nil {
		itf.NewOwner1 = tx.Output1.Owner
		itf.Denom1 = NewUint256(tx.Output1.Denom)
	} else {
		itf.Denom1 = NewUint256(nil)
	}
	itf.Fee = NewUint256(tx.Fee)

	return rlp.Encode(w, &itf)
}

func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	var itf rlpTransactionHelper
	err := s.Decode(&itf)
	if err != nil {
		return err
	}
	tx.Input0 = NewInput(
		util.Big2Uint64(itf.BlkNum0.ToBig()),
		util.Big2Uint32(itf.TxIdx0.ToBig()),
		util.Big2Uint8(itf.OutIdx0.ToBig()),
		itf.DepositNonce0.ToBig(),
		itf.Owner0,
	)
	tx.Input1 = NewInput(
		util.Big2Uint64(itf.BlkNum1.ToBig()),
		util.Big2Uint32(itf.TxIdx1.ToBig()),
		util.Big2Uint8(itf.OutIdx1.ToBig()),
		itf.DepositNonce1.ToBig(),
		itf.Owner1,
	)
	tx.Output0 = NewOutput(itf.NewOwner0, itf.Denom0.ToBig(), Zero())
	tx.Output1 = NewOutput(itf.NewOwner1, itf.Denom1.ToBig(), Zero())
	tx.Sig0 = itf.Sig0
	tx.Sig1 = itf.Sig1
	tx.Fee = itf.Fee.ToBig()

	return nil
}
