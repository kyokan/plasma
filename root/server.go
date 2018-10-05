package root

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/plasma/chain"
	"github.com/kyokan/plasma/db"
	"github.com/kyokan/plasma/rpc"
	"github.com/kyokan/plasma/rpc/pb"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct {
	storage db.PlasmaStorage
	ctx     context.Context
}

func NewServer(ctx context.Context, storage db.PlasmaStorage) (*Server) {
	return &Server{
		storage: storage,
		ctx:     ctx,
	}
}

func (r *Server) Start(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Println("error", err)
		return err
	}

	s := grpc.NewServer()
	pb.RegisterRootServer(s, r)
	if err := s.Serve(lis); err != nil {
		log.Println("error", err)
		return err
	}

	go func() {
		<-r.ctx.Done()
		s.Stop()
	}()

	return nil
}

func (r *Server) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	addr := common.BytesToAddress(req.Address)
	bal, err := r.storage.Balance(&addr)
	if err != nil {
		return nil, err
	}

	return &pb.GetBalanceResponse{
		Balance: rpc.SerializeBig(bal),
	}, nil
}

func (r *Server) GetUTXOs(ctx context.Context, req *pb.GetUTXOsRequest) (*pb.GetUTXOsResponse, error) {
	addr := common.BytesToAddress(req.Address)
	var txs []chain.Transaction
	var err error
	if req.Spendable {
		txs, err = r.storage.SpendableTxs(&addr)
	} else {
		txs, err = r.storage.UTXOs(&addr)
	}

	if err != nil {
		return nil, err
	}

	return &pb.GetUTXOsResponse{
		Transactions: rpc.SerializeTxs(txs),
	}, nil
}

func (r *Server) GetBlock(ctx context.Context, req *pb.GetBlockRequest) (*pb.GetBlockResponse, error) {
	block, err := r.storage.BlockAtHeight(req.Number)
	if err != nil {
		return nil, err
	}

	txs, err := r.storage.FindTransactionsByBlockNum(block.Header.Number)
	if err != nil {
		return nil, err
	}

	res := &pb.GetBlockResponse{
		Block: &pb.Block{
			Header: &pb.BlockHeader{
				MerkleRoot:    block.Header.MerkleRoot,
				RlpMerkleRoot: block.Header.RLPMerkleRoot,
				PrevHash:      block.Header.PrevHash,
				Number:        req.Number,
			},
			Hash: block.BlockHash,
		},
		Transactions: rpc.SerializeTxs(txs),
	}

	return res, nil
}

func (r *Server) Send(ctx context.Context, req *pb.SendRequest) (*pb.SendResponse, error) {
	tx := rpc.DeserializeTx(req.Transaction)

	signedTx, storeErr := r.storage.StoreTransaction(*tx)
	if storeErr != nil {
		return nil, storeErr
	}
	return &pb.SendResponse{
		Transaction: rpc.SerializeTx(signedTx),
	}, nil
}
