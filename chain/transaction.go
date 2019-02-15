package chain

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/util"
	"github.com/kyokan/plasma/rpc/pb"
	"github.com/pkg/errors"
)

type Transaction struct {
	Body *TransactionBody
	Sigs [2]Signature
}

type rlpConfirmedTransaction struct {
	Transaction rlpTransactionBody
	Body        [2]Signature
}

func (c *Transaction) RLPHash(hasher util.Hasher) util.Hash {
	bytes := c.RLP()
	return hasher(bytes)
}

func (c *Transaction) RLP() []byte {
	bytes, err := rlp.EncodeToBytes(rlpConfirmedTransaction{
		Transaction: c.Body.rlpRepresentation(),
		Body:        c.Sigs,
	})
	if err != nil {
		panic(err)
	}

	return bytes
}

func (c *Transaction) Proto() (*pb.Transaction) {
	sig0 := make([]byte, len(c.Sigs[0]), len(c.Sigs[0]))
	copy(sig0, c.Sigs[0][:])
	sig1 := make([]byte, len(c.Sigs[1]), len(c.Sigs[1]))
	copy(sig1, c.Sigs[1][:])

	return &pb.Transaction{
		Body: c.Body.Proto(),
		Sig0: sig0,
		Sig1: sig1,
	}
}

func (c *Transaction) Clone() (*Transaction) {
	proto := c.Proto()
	clone, err := TransactionFromProto(proto)
	if err != nil {
		// should never happen
		panic(err)
	}

	return clone
}

func TransactionFromProto(protoTx *pb.Transaction) (*Transaction, error) {
	if protoTx == nil {
		return nil, errors.New("proto tx cannot be nil")
	}

	body, err := TransactionBodyFromProto(protoTx.Body)
	if err != nil {
		return nil, err
	}

	var sig0 Signature
	copy(sig0[:], protoTx.Sig0)
	var sig1 Signature
	copy(sig1[:], protoTx.Sig0)

	return &Transaction{
		Body: body,
		Sigs: [2]Signature{
			sig0,
			sig1,
		},
	}, nil
}
