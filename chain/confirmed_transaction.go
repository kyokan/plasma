package chain

import (
	"github.com/kyokan/plasma/rpc/pb"
	"github.com/kyokan/plasma/util"
	"github.com/golang/protobuf/proto"
)

type ConfirmedTransaction struct {
	Transaction *Transaction
	ConfirmSigs [2]Signature
}

var zeroSig Signature

func (c *ConfirmedTransaction) IsConfirmed() bool {
	return c.ConfirmSigs[0] != zeroSig &&
			c.ConfirmSigs[1] != zeroSig
}

func (c *ConfirmedTransaction) Proto() (*pb.ConfirmedTransaction) {
	sig0 := make([]byte, len(c.ConfirmSigs[0]), len(c.ConfirmSigs[0]))
	copy(sig0, c.ConfirmSigs[0][:])
	sig1 := make([]byte, len(c.ConfirmSigs[1]), len(c.ConfirmSigs[1]))
	copy(sig1, c.ConfirmSigs[1][:])

	return &pb.ConfirmedTransaction{
		Transaction: c.Transaction.Proto(),
		ConfirmSig0: sig0,
		ConfirmSig1: sig1,
	}
}

func (c *ConfirmedTransaction) MarshalProto() ([]byte, error) {
	return proto.Marshal(c.Proto())
}

func (c *ConfirmedTransaction) Hash() util.Hash {
	return c.Transaction.RLPHash(util.Sha256)
}

func ConfirmedTransactionFromProto(protoTx *pb.ConfirmedTransaction) (*ConfirmedTransaction, error) {
	tx, err := TransactionFromProto(protoTx.Transaction)
	if err != nil {
		return nil, err
	}

	var c ConfirmedTransaction
	c.Transaction = tx
	copy(c.ConfirmSigs[0][:], protoTx.ConfirmSig0)
	copy(c.ConfirmSigs[1][:], protoTx.ConfirmSig1)
	return &c, nil
}

func UnmarshalConfirmedTransactionProto(b []byte) (*ConfirmedTransaction, error) {
	var protoTx pb.ConfirmedTransaction
	err := proto.Unmarshal(b, &protoTx)
	if err != nil {
		return nil, err
	}
	return ConfirmedTransactionFromProto(&protoTx)
}