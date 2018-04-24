package rpc

import (
	"log"
	"net/http"

	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
)

type GetBlocksArgs struct {
	Height uint64
}

type GetBlocksResponse struct {
	Block        *chain.Block
	Transactions []chain.Transaction
}

type BlockService struct {
	DB *db.Database
}

func (t *BlockService) GetBlock(r *http.Request, args *GetBlocksArgs, reply *GetBlocksResponse) error {
	log.Println("Received Block.GetBlocks request.")

	height := args.Height

	block, err := t.DB.BlockDao.BlockAtHeight(height)

	if err != nil {
		return err
	}

	txs, err := t.DB.TxDao.FindByBlockNum(height)

	if err != nil {
		return err
	}

	*reply = GetBlocksResponse{
		Block:        block,
		Transactions: txs,
	}

	return nil
}
