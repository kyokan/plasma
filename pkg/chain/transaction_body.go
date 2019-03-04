package chain

import (
	"bytes"
	"fmt"
	"math/big"

	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/pkg/rpc"
	"github.com/kyokan/plasma/pkg/rpc/pb"
	"github.com/kyokan/plasma/util"
)

type TransactionBody struct {
	Input0            *Input
	Input0ConfirmSigs [2]Signature
	Input1            *Input
	Input1ConfirmSigs [2]Signature
	Output0           *Output
	Output1           *Output
	Fee               *big.Int
	BlockNumber       uint64
	TransactionIndex  uint32
}

type rlpTransactionBody struct {
	BlkNum0           *UInt256
	TxIdx0            *UInt256
	OutIdx0           *UInt256
	DepositNonce0     *UInt256
	Input0ConfirmSigs [130]byte
	BlkNum1           *UInt256
	TxIdx1            *UInt256
	OutIdx1           *UInt256
	DepositNonce1     *UInt256
	Input1ConfirmSigs [130]byte
	Owner0            common.Address
	Amount0           *UInt256
	Owner1            common.Address
	Amount1           *UInt256
	Fee               *UInt256
}

type transactionBodyJSON struct {
	Input0            *Input       `json:"input0"`
	Input0ConfirmSigs [2]Signature `json:"input0ConfirmSigs"`
	Input1            *Input       `json:"input1"`
	Input1ConfirmSigs [2]Signature `json:"input1ConfirmSigs"`
	Output0           *Output      `json:"output0"`
	Output1           *Output      `json:"output1"`
	Fee               string       `json:"fee"`
	BlockNumber       uint64       `json:"blockNumber"`
	TransactionIndex  uint32       `json:"transactionIndex"`
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
		Input0:            b.Input0,
		Input0ConfirmSigs: b.Input0ConfirmSigs,
		Input1:            b.Input1,
		Input1ConfirmSigs: b.Input1ConfirmSigs,
		Output0:           b.Output0,
		Output1:           b.Output1,
		Fee:               util.Big2Str(b.Fee),
		BlockNumber:       b.BlockNumber,
		TransactionIndex:  b.TransactionIndex,
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
	b.Input0ConfirmSigs = jsonRep.Input0ConfirmSigs
	b.Input1 = jsonRep.Input1
	b.Input1ConfirmSigs = jsonRep.Input1ConfirmSigs
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

func (b *TransactionBody) lookupOutput(addr *common.Address) ([]*Output, []uint8) {
	var outputs []*Output
	var indices []uint8
	output := b.OutputAt(0)

	if output.Owner == *addr {
		outputs = append(outputs, output)
		indices = append(indices, 0)
	}

	output = b.OutputAt(1)
	if output.Owner == *addr {
		outputs = append(outputs, output)
		indices = append(indices, 1)
	}

	if len(outputs) == 0 {
		panic(fmt.Sprint("No output found for address: ", addr.Hex()))
	}

	return outputs, indices
}

func (b *TransactionBody) OutputsFor(addr *common.Address) []*Output {
	outputs, _ := b.lookupOutput(addr)
	return outputs
}

func (b *TransactionBody) OutputIndicesFor(addr *common.Address) []uint8 {
	_, idx := b.lookupOutput(addr)
	return idx
}

func (b *TransactionBody) rlpRepresentation() rlpTransactionBody {
	var input0ConfirmSigs [130]byte
	var input0SigBuf bytes.Buffer
	input0SigBuf.Write(b.Input0ConfirmSigs[0][:])
	input0SigBuf.Write(b.Input0ConfirmSigs[1][:])
	copy(input0ConfirmSigs[:], input0SigBuf.Bytes())

	var input1ConfirmSigs [130]byte
	var input1SigBuf bytes.Buffer
	input1SigBuf.Write(b.Input1ConfirmSigs[0][:])
	input1SigBuf.Write(b.Input1ConfirmSigs[1][:])
	copy(input1ConfirmSigs[:], input1SigBuf.Bytes())

	return rlpTransactionBody{
		BlkNum0:           NewUint256(util.Uint642Big(b.Input0.BlockNumber)),
		TxIdx0:            NewUint256(util.Uint322Big(b.Input0.TransactionIndex)),
		OutIdx0:           NewUint256(util.Uint82Big(b.Input0.OutputIndex)),
		DepositNonce0:     NewUint256(b.Input0.DepositNonce),
		Input0ConfirmSigs: input0ConfirmSigs,
		BlkNum1:           NewUint256(util.Uint642Big(b.Input1.BlockNumber)),
		TxIdx1:            NewUint256(util.Uint322Big(b.Input1.TransactionIndex)),
		OutIdx1:           NewUint256(util.Uint82Big(b.Input1.OutputIndex)),
		DepositNonce1:     NewUint256(b.Input1.DepositNonce),
		Input1ConfirmSigs: input1ConfirmSigs,
		Owner0:            b.Output0.Owner,
		Amount0:           NewUint256(b.Output0.Amount),
		Owner1:            b.Output1.Owner,
		Amount1:           NewUint256(b.Output1.Amount),
		Fee:               NewUint256(b.Fee),
	}
}

func (b *TransactionBody) SignatureHash() util.Hash {
	return b.RLPHash(util.Keccak256)
}

func (b *TransactionBody) RLP() []byte {
	buf, err := rlp.EncodeToBytes(b.rlpRepresentation())
	if err != nil {
		panic(err)
	}

	return buf
}

func (b *TransactionBody) RLPHash(hasher util.Hasher) util.Hash {
	return hasher(b.RLP())
}

func (b *TransactionBody) Proto() *pb.TransactionBody {
	input0ConfirmSig0 := make([]byte, 65, 65)
	copy(input0ConfirmSig0, b.Input0ConfirmSigs[0][:])
	input0ConfirmSig1 := make([]byte, 65, 65)
	copy(input0ConfirmSig1, b.Input0ConfirmSigs[1][:])
	input1ConfirmSig0 := make([]byte, 65, 65)
	copy(input1ConfirmSig0, b.Input1ConfirmSigs[0][:])
	input1ConfirmSig1 := make([]byte, 65, 65)
	copy(input1ConfirmSig1, b.Input1ConfirmSigs[1][:])

	return &pb.TransactionBody{
		Input0:            b.Input0.Proto(),
		Input0ConfirmSig0: input0ConfirmSig0,
		Input0ConfirmSig1: input0ConfirmSig1,
		Input1:            b.Input1.Proto(),
		Input1ConfirmSig0: input1ConfirmSig0,
		Input1ConfirmSig1: input1ConfirmSig1,
		Output0:           b.Output0.Proto(),
		Output1:           b.Output1.Proto(),
		Fee:               rpc.SerializeBig(b.Fee),
		BlockNum:          b.BlockNumber,
		TxIdx:             b.TransactionIndex,
	}
}

func TransactionBodyFromProto(protoBody *pb.TransactionBody) (*TransactionBody, error) {
	input0, err := InputFromProto(protoBody.Input0)
	if err != nil {
		return nil, err
	}
	var input0ConfirmSigs [2]Signature
	copy(input0ConfirmSigs[0][:], protoBody.Input0ConfirmSig0)
	copy(input0ConfirmSigs[1][:], protoBody.Input0ConfirmSig1)

	input1, err := InputFromProto(protoBody.Input1)
	if err != nil {
		return nil, err
	}
	var input1ConfirmSigs [2]Signature
	copy(input1ConfirmSigs[0][:], protoBody.Input1ConfirmSig0)
	copy(input1ConfirmSigs[1][:], protoBody.Input1ConfirmSig1)

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
		Input0:            input0,
		Input0ConfirmSigs: input0ConfirmSigs,
		Input1:            input1,
		Input1ConfirmSigs: input1ConfirmSigs,
		Output0:           output0,
		Output1:           output1,
		Fee:               fee,
		BlockNumber:       blockNum,
		TransactionIndex:  txIdx,
	}, nil
}
