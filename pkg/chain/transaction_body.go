package chain

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/util"
	"github.com/kyokan/plasma/pkg/rpc/pb"
	"github.com/kyokan/plasma/pkg/rpc"
	"encoding/json"
)

type TransactionBody struct {
	Input0           *Input
	Input0ConfirmSig Signature
	Input1           *Input
	Input1ConfirmSig Signature
	Output0          *Output
	Output1          *Output
	Fee              *big.Int
	BlockNumber      uint64
	TransactionIndex uint32
}

type rlpTransactionBody struct {
	BlkNum0          *UInt256
	TxIdx0           *UInt256
	OutIdx0          *UInt256
	DepositNonce0    *UInt256
	Input0ConfirmSig Signature
	BlkNum1          *UInt256
	TxIdx1           *UInt256
	OutIdx1          *UInt256
	DepositNonce1    *UInt256
	Input1ConfirmSig Signature
	Owner0           common.Address
	Amount0          *UInt256
	Owner1           common.Address
	Amount1          *UInt256
	Fee              *UInt256
}

type transactionBodyJSON struct {
	Input0           *Input    `json:"input0"`
	Input0ConfirmSig Signature `json:"input0ConfirmSig"`
	Input1           *Input    `json:"input1"`
	Input1ConfirmSig Signature `json:"input1ConfirmSig"`
	Output0          *Output   `json:"output0"`
	Output1          *Output   `json:"output1"`
	Fee              string    `json:"fee"`
	BlockNumber      uint64    `json:"blockNumber"`
	TransactionIndex uint32    `json:"transactionIndex"`
}

func ZeroBody() *TransactionBody {
	return &TransactionBody{
		Input0:  ZeroInput(),
		Input1:  ZeroInput(),
		Output0: ZeroOutput(),
		Output1: ZeroOutput(),
		Fee:     Zero(),
	}
}

func (b *TransactionBody) MarshalJSON() ([]byte, error) {
	jsonRep := &transactionBodyJSON{
		Input0:           b.Input0,
		Input0ConfirmSig: b.Input0ConfirmSig,
		Input1:           b.Input1,
		Input1ConfirmSig: b.Input1ConfirmSig,
		Output0:          b.Output0,
		Output1:          b.Output1,
		Fee:              util.Big2Str(b.Fee),
		BlockNumber:      b.BlockNumber,
		TransactionIndex: b.TransactionIndex,
	}
	return json.Marshal(jsonRep)
}

func (b *TransactionBody) UnmarshalJSON(in []byte) error {
	jsonRep := &transactionBodyJSON{}
	err := json.Unmarshal(in, &jsonRep)
	if err != nil {
		return err
	}
	fee, err := util.Str2Big(jsonRep.Fee)
	if err != nil {
		return err
	}
	b.Input0 = jsonRep.Input0
	b.Input0ConfirmSig = jsonRep.Input0ConfirmSig
	b.Input1 = jsonRep.Input1
	b.Input1ConfirmSig = jsonRep.Input1ConfirmSig
	b.Output0 = jsonRep.Output0
	b.Output1 = jsonRep.Output1
	b.Fee = fee
	b.BlockNumber = jsonRep.BlockNumber
	b.TransactionIndex = jsonRep.TransactionIndex
	return nil
}

func (b *TransactionBody) IsDeposit() bool {
	return b.Input0.IsDeposit()
}

func (b *TransactionBody) IsZero() bool {
	if b.IsDeposit() {
		return false
	}
	return b.Input0.IsZero() &&
			b.Input1.IsZero() &&
			b.Output0.IsZeroOutput() &&
			b.Output1.IsZeroOutput()
}

func (b *TransactionBody) InputAt(idx uint8) *Input {
	if idx != 0 && idx != 1 {
		panic(fmt.Sprint("Invalid input index: ", idx))
	}

	if idx == 0 {
		return b.Input0
	}

	return b.Input1
}

func (b *TransactionBody) OutputAt(idx uint8) *Output {
	if idx == 0 {
		return b.Output0
	}

	return b.Output1
}

func (b *TransactionBody) lookupOutput(addr *common.Address) (*Output, uint8) {
	output := b.OutputAt(0)

	if util.AddressesEqual(&output.Owner, addr) {
		return output, 0
	}

	output = b.OutputAt(1)

	if util.AddressesEqual(&output.Owner, addr) {
		return output, 1
	}

	panic(fmt.Sprint("No output found for address: ", addr.Hex()))
}

func (b *TransactionBody) OutputFor(addr *common.Address) *Output {
	out, _ := b.lookupOutput(addr)
	return out
}

func (b *TransactionBody) OutputIndexFor(addr *common.Address) uint8 {
	_, idx := b.lookupOutput(addr)
	return idx
}

func (b *TransactionBody) rlpRepresentation() rlpTransactionBody {
	return rlpTransactionBody{
		BlkNum0:          NewUint256(util.Uint642Big(b.Input0.BlockNumber)),
		TxIdx0:           NewUint256(util.Uint322Big(b.Input0.TransactionIndex)),
		OutIdx0:          NewUint256(util.Uint82Big(b.Input0.OutputIndex)),
		DepositNonce0:    NewUint256(b.Input0.DepositNonce),
		Input0ConfirmSig: b.Input0ConfirmSig,
		BlkNum1:          NewUint256(util.Uint642Big(b.Input1.BlockNumber)),
		TxIdx1:           NewUint256(util.Uint322Big(b.Input1.TransactionIndex)),
		OutIdx1:          NewUint256(util.Uint82Big(b.Input1.OutputIndex)),
		DepositNonce1:    NewUint256(b.Input1.DepositNonce),
		Input1ConfirmSig: b.Input1ConfirmSig,
		Owner0:           b.Output0.Owner,
		Amount0:          NewUint256(b.Output0.Amount),
		Owner1:           b.Output1.Owner,
		Amount1:          NewUint256(b.Output1.Amount),
		Fee:              NewUint256(b.Fee),
	}
}

func (b *TransactionBody) SignatureHash() util.Hash {
	return b.RLPHash(util.Keccak256)
}

func (b *TransactionBody) RLP() []byte {
	bytes, err := rlp.EncodeToBytes(b.rlpRepresentation())
	if err != nil {
		panic(err)
	}

	return bytes
}

func (b *TransactionBody) RLPHash(hasher util.Hasher) util.Hash {
	return hasher(b.RLP())
}

func (b *TransactionBody) Proto() (*pb.TransactionBody) {
	confirmSig0 := make([]byte, len(b.Input0ConfirmSig), len(b.Input0ConfirmSig))
	copy(confirmSig0, b.Input0ConfirmSig[:])
	confirmSig1 := make([]byte, len(b.Input1ConfirmSig), len(b.Input1ConfirmSig))
	copy(confirmSig1, b.Input1ConfirmSig[:])

	return &pb.TransactionBody{
		Input0:           b.Input0.Proto(),
		Input0ConfirmSig: confirmSig0,
		Input1:           b.Input1.Proto(),
		Input1ConfirmSig: confirmSig1,
		Output0:          b.Output0.Proto(),
		Output1:          b.Output1.Proto(),
		Fee:              rpc.SerializeBig(b.Fee),
		BlockNum:         b.BlockNumber,
		TxIdx:            b.TransactionIndex,
	}
}

func TransactionBodyFromProto(protoBody *pb.TransactionBody) (*TransactionBody, error) {
	input0, err := InputFromProto(protoBody.Input0)
	if err != nil {
		return nil, err
	}
	var input0ConfirmSig Signature
	copy(input0ConfirmSig[:], protoBody.Input0ConfirmSig)

	input1, err := InputFromProto(protoBody.Input1)
	if err != nil {
		return nil, err
	}
	var input1ConfirmSig Signature
	copy(input1ConfirmSig[:], protoBody.Input1ConfirmSig)

	output0, err := OutputFromProto(protoBody.Output0)
	if err != nil {
		return nil, err
	}
	output1, err := OutputFromProto(protoBody.Output1)
	if err != nil {
		return nil, err
	}

	fee := rpc.DeserializeBig(protoBody.Fee)
	blockNum := protoBody.BlockNum
	txIdx := protoBody.TxIdx

	return &TransactionBody{
		Input0:           input0,
		Input0ConfirmSig: input0ConfirmSig,
		Input1:           input1,
		Input1ConfirmSig: input1ConfirmSig,
		Output0:          output0,
		Output1:          output1,
		Fee:              fee,
		BlockNumber:      blockNum,
		TransactionIndex: txIdx,
	}, nil
}
