package chain

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"github.com/kyokan/plasma/rpc/pb"
	"github.com/kyokan/plasma/rpc"
)

type Output struct {
	Owner  common.Address
	Amount *big.Int
}

func NewOutput(newOwner common.Address, amount *big.Int) *Output {
	return &Output{
		Owner:  common.BytesToAddress(newOwner.Bytes()),
		Amount: big.NewInt(amount.Int64()),
	}
}

func ZeroOutput() *Output {
	return &Output{
		Owner:  common.BytesToAddress(make([]byte, 20, 20)),
		Amount: big.NewInt(0),
	}
}

func (out *Output) IsExit() bool {
	if out == nil {
		return false
	}
	exit := ExitOutput()
	for i := 0; i != len(out.Owner); i++ {
		if out.Owner[i] != exit.Owner[i] {
			return false
		}
	}
	return true
}

func (out *Output) IsZeroOutput() bool {
	if out == nil {
		return true
	}
	addrBytes := out.Owner.Bytes()

	for _, v := range addrBytes {
		if v != 0 {
			return false
		}
	}

	return out.Amount == nil || out.Amount.Cmp(Zero()) == 0
}

func (out *Output) Proto() (*pb.Output) {
	owner := make([]byte, len(out.Owner), len(out.Owner))
	copy(owner, out.Owner[:])

	return &pb.Output{
		Owner:  owner,
		Amount: rpc.SerializeBig(out.Amount),
	}
}

func OutputFromProto(outProto *pb.Output) (*Output, error) {
	out := &Output{}
	var owner common.Address
	copy(owner[:], outProto.Owner)
	amount := rpc.DeserializeBig(outProto.Amount)
	out.Owner = owner
	out.Amount = amount
	return out, nil
}
