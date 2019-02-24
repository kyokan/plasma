package validation

import (
	"fmt"
	"math/big"
)

type ErrNegativeOutput struct {
	Index uint8
}

func NewErrNegativeOutput(index uint8) error {
	return &ErrNegativeOutput{
		Index: index,
	}
}

func (e *ErrNegativeOutput) Error() string {
	return fmt.Sprintf("output %d is negative", e.Index)
}

type ErrTxNotFound struct {
	InputIndex       uint8
	BlockNumber      uint64
	TransactionIndex uint32
}

func NewErrTxNotFound(inputIndex uint8, blockNum uint64, txIdx uint32) error {
	return &ErrTxNotFound{
		InputIndex:       inputIndex,
		BlockNumber:      blockNum,
		TransactionIndex: txIdx,
	}
}

func (e *ErrTxNotFound) Error() string {
	return fmt.Sprintf("transaction in block %d with index %d for input %d not found", e.BlockNumber, e.TransactionIndex, e.InputIndex)
}

type ErrConfirmSigMismatch struct {
	InputIndex uint8
}

func NewErrConfirmSigMismatch(inputIndex uint8) error {
	return &ErrConfirmSigMismatch{
		InputIndex: inputIndex,
	}
}

func (e *ErrConfirmSigMismatch) Error() string {
	return fmt.Sprintf("mismatched confirm sigs for input %d", e.InputIndex)
}

type ErrInvalidSignature struct {
	InputIndex uint8
}

func NewErrInvalidSignature(inputIndex uint8) error {
	return &ErrInvalidSignature{
		InputIndex: inputIndex,
	}
}

type ErrIdenticalInputs struct {
}

func NewErrIdenticalInputs() error {
	return &ErrIdenticalInputs{}
}

func (e *ErrIdenticalInputs) Error() string {
	return "input0 and input1 spend identical outputs"
}

func (e *ErrInvalidSignature) Error() string {
	return fmt.Sprintf("invalid signature for input %d", e.InputIndex)
}

type ErrInputOutputValueMismatch struct {
	TotalInputs      *big.Int
	TotalOutputsFees *big.Int
}

func NewErrInputOutputValueMismatch(totalInputs *big.Int, totalOutputsFees *big.Int) error {
	return &ErrInputOutputValueMismatch{
		TotalInputs:      totalInputs,
		TotalOutputsFees: totalOutputsFees,
	}
}

func (e *ErrInputOutputValueMismatch) Error() string {
	return fmt.Sprintf(
		"total inputs of %s do not match total outputs and fees of %s",
		e.TotalInputs.Text(10),
		e.TotalOutputsFees.Text(10),
	)
}

type ErrDoubleSpent struct{}

func NewErrDoubleSpent() error {
	return &ErrDoubleSpent{}
}

func (e *ErrDoubleSpent) Error() string {
	return "transaction double spent"
}

type ErrDepositDefinedInput1 struct{}

func NewErrDepositDefinedInput1() error {
	return &ErrDepositDefinedInput1{}
}

func (e *ErrDepositDefinedInput1) Error() string {
	return "deposit defined input1, which is illegal"
}

type ErrDepositNonEmptyConfirmSig struct {}

func NewErrDepositNonEmptyConfirmSig() error {
	return &ErrDepositNonEmptyConfirmSig{}
}

func (e *ErrDepositNonEmptyConfirmSig) Error() string {
	return "deposit defined non-empty input0 confirm sig, which is illegal"
}