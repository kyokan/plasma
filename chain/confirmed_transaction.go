package chain

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kyokan/plasma/util"
				)

type ConfirmedTransaction struct {
	Transaction Transaction
	Signatures  [2]Signature
}

type rlpConfirmedTransaction struct {
	Transaction rlpTransaction
	Signatures [2]Signature
}

func (c *ConfirmedTransaction) RLPHash(hasher util.Hasher) util.Hash {
	bytes, err := rlp.EncodeToBytes(rlpConfirmedTransaction{
		Transaction: c.Transaction.rlpRepresentation(),
		Signatures: c.Signatures,
	})

	if err != nil {
		panic(err)
	}

	return hasher(bytes)
}