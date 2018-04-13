package node

import (
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
)

type BlockSink struct {
	c      chan chain.Block
	db     *db.Database
	client *eth.Client
}

type BlockRequest struct {
	Start  int
	Offset int
}

type BlockResponse struct {
	Error error
	Block *chain.Block
}

func NewBlockSink(db *db.Database, client *eth.Client) *BlockSink {
	return &BlockSink{c: make(chan chain.Block), db: db, client: client}
}
