package rpc

import (
	"log"
	"net/http"

	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/eth"
)

type GetBlocksArgs struct {
	Start  int
	Offset int
}

type GetBlocksResponse struct {
	Block       *chain.Block
	Transaction []chain.Transaction
}

type BlockService struct {
	db     *db.Database
	client *eth.Client
}

func (t *BlockService) GetBlocks(r *http.Request, args *GetBlocksArgs, reply *GetBlocksResponse) error {
	log.Printf("Received Block.GetBlocks request.")

	start := args.Start
	offset := args.Offset

	// req := node.BlockRequest{
	// 	Start:  start,
	// 	Offset: offset,
	// }

	// ch := make(chan node.BlockRequest)
	// t.BlockChan <- ch
	// ch <- req
	// // res := <-ch
	// close(ch)
	// // Query db.

	// node.DB.BlockDao.Save(nil)

	*reply = GetBlocksResponse{
		Block:       nil,
		Transaction: nil,
	}

	return nil
}
