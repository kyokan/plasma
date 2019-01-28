package chain

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/util"
	"math/big"
		)

type ConfirmedTransaction struct {
	Transaction Transaction
	Signatures  [2]Signature
}

func (c *ConfirmedTransaction) RLPHash(hasher util.Hasher) util.Hash {
	bytes, err := rlp.EncodeToBytes(c)

	if err != nil {
		panic(err)
	}

	return hasher(bytes)
}

func (c *ConfirmedTransaction) Hash(hasher util.Hasher) util.Hash {
	return c.RLPHash(hasher)
}

func (c *ConfirmedTransaction) GetFee() *big.Int {
	return c.Transaction.GetFee()
}

func (c *ConfirmedTransaction) SetIndex(idx uint32) {
	c.Transaction.SetIndex(idx)
}
